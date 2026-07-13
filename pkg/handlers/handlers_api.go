package handlers

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"otaku-quiz-africa/pkg/database"
)

func (h *Handler) APIQuizVote(c *fiber.Ctx) error {
	type Request struct {
		QuizID   string `json:"quiz_id"`
		VoteType string `json:"vote_type"`
	}
	var req Request
	c.BodyParser(&req)

	sess, _ := h.store.Get(c)
	userID := sess.Get("user_id").(string)

	if req.QuizID == "" || (req.VoteType != "like" && req.VoteType != "dislike") {
		return c.Status(400).JSON(fiber.Map{"error": "Données invalides"})
	}

	checkBody, _ := h.db.Select("quiz_votes",
		fmt.Sprintf("quiz_id=eq.%s&user_id=eq.%s&select=id,vote_type", req.QuizID, userID), true)
	var existing []map[string]interface{}
	json.Unmarshal(checkBody, &existing)

	if len(existing) > 0 {
		oldVote, _ := existing[0]["vote_type"].(string)
		if oldVote == req.VoteType {
			h.db.Delete("quiz_votes", fmt.Sprintf("quiz_id=eq.%s&user_id=eq.%s", req.QuizID, userID), true)
		} else {
			h.db.Update("quiz_votes",
				fmt.Sprintf("quiz_id=eq.%s&user_id=eq.%s", req.QuizID, userID),
				[]byte(fmt.Sprintf(`{"vote_type":"%s"}`, req.VoteType)), true)
		}
	} else {
		voteData, _ := json.Marshal(map[string]interface{}{
			"quiz_id":   req.QuizID,
			"user_id":   userID,
			"vote_type": req.VoteType,
		})
		h.db.Insert("quiz_votes", voteData, true)
	}

	likeBody, _ := h.db.Select("quiz_votes",
		fmt.Sprintf("quiz_id=eq.%s&vote_type=eq.like&select=id", req.QuizID), true)
	var likes []map[string]interface{}
	json.Unmarshal(likeBody, &likes)
	likeCount := len(likes)

	dislikeBody, _ := h.db.Select("quiz_votes",
		fmt.Sprintf("quiz_id=eq.%s&vote_type=eq.dislike&select=id", req.QuizID), true)
	var dislikes []map[string]interface{}
	json.Unmarshal(dislikeBody, &dislikes)
	dislikeCount := len(dislikes)

	updateData, _ := json.Marshal(map[string]interface{}{"like_count": likeCount, "dislike_count": dislikeCount})
	h.db.Update("quizzes", fmt.Sprintf("id=eq.%s", req.QuizID), updateData, true)

	return c.JSON(fiber.Map{"success": true, "like_count": likeCount, "dislike_count": dislikeCount})
}

func (h *Handler) APIQuizUploadImage(c *fiber.Ctx) error {
	user := c.Locals("user").(*UserProfile)

	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Aucun fichier"})
	}

	f, err := file.Open()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Erreur lecture fichier"})
	}
	defer f.Close()
	fileBytes, _ := io.ReadAll(f)

	cloudName := os.Getenv("NEXT_PUBLIC_CLOUDINARY_CLOUD_NAME")
	apiKey := os.Getenv("CLOUDINARY_API_KEY")
	apiSecret := os.Getenv("CLOUDINARY_API_SECRET")

	if cloudName == "" || apiKey == "" || apiSecret == "" {
		return c.Status(500).JSON(fiber.Map{"error": "Cloudinary non configuré"})
	}

	timestamp := time.Now().Unix()
	paramsToSign := fmt.Sprintf("folder=avatars&public_id=%s&timestamp=%d", user.ID, timestamp)
	hasher := sha1.New()
	hasher.Write([]byte(paramsToSign + apiSecret))
	signature := fmt.Sprintf("%x", hasher.Sum(nil))

	uploadURL := fmt.Sprintf("https://api.cloudinary.com/v1_1/%s/image/upload", cloudName)

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, _ := writer.CreateFormFile("file", file.Filename)
	part.Write(fileBytes)

	writer.WriteField("api_key", apiKey)
	writer.WriteField("timestamp", fmt.Sprintf("%d", timestamp))
	writer.WriteField("signature", signature)
	writer.WriteField("folder", "avatars")
	writer.WriteField("public_id", user.ID)

	writer.Close()

	httpReq, _ := http.NewRequest("POST", uploadURL, &buf)
	httpReq.Header.Set("Content-Type", writer.FormDataContentType())

	httpResp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Erreur upload Cloudinary: " + err.Error()})
	}
	defer httpResp.Body.Close()

	respBody, _ := io.ReadAll(httpResp.Body)

	if httpResp.StatusCode >= 400 {
		return c.Status(500).JSON(fiber.Map{"error": "Erreur Cloudinary: " + string(respBody)})
	}

	var result map[string]interface{}
	json.Unmarshal(respBody, &result)
	publicURL, _ := result["secure_url"].(string)

	if publicURL == "" {
		return c.Status(500).JSON(fiber.Map{"error": "URL non retournée par Cloudinary"})
	}

	updateData, _ := json.Marshal(map[string]interface{}{
		"avatar_url": publicURL,
	})
	h.db.Update("user_profiles", fmt.Sprintf("id=eq.%s", user.ID), updateData, true)

	return c.JSON(fiber.Map{"success": true, "url": publicURL})
}

func (h *Handler) APIGetChallenge(c *fiber.Ctx) error {
	challengeID := c.Params("id")
	if challengeID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "ID requis"})
	}

	body, err := h.db.Select("challenge_sessions",
		fmt.Sprintf("id=eq.%s&select=*", challengeID), true)
	if err != nil {
		log.Printf("[APIGetChallenge] Select error: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	var sessions []map[string]interface{}
	json.Unmarshal(body, &sessions)

	if len(sessions) == 0 {
		log.Printf("[APIGetChallenge] No session found for id=%s body=%s", challengeID, string(body)[:min(len(body), 200)])
		return c.Status(404).JSON(fiber.Map{"error": "Défi introuvable"})
	}

	session := sessions[0]
	quizID := database.DBValue(session["quiz_id"])

	quizTitle := "Quiz"
	questionCount := 0
	qBody, _ := h.db.Select("quizzes",
		fmt.Sprintf("id=eq.%s&select=title,question_count", quizID), true)
	var qRows []map[string]interface{}
	json.Unmarshal(qBody, &qRows)
	if len(qRows) > 0 {
		quizTitle, _ = qRows[0]["title"].(string)
		questionCount = database.DBInt(qRows[0]["question_count"])
	}

	creatorID := database.DBValue(session["creator_id"])
	creatorName := "Créateur"
	cBody, _ := h.db.Select("user_profiles",
		fmt.Sprintf("id=eq.%s&select=username,nickname", creatorID), false)
	var cRows []map[string]interface{}
	json.Unmarshal(cBody, &cRows)
	if len(cRows) > 0 {
		if nn, ok := cRows[0]["nickname"].(string); ok && nn != "" {
			creatorName = nn
		} else if un, ok := cRows[0]["username"].(string); ok {
			creatorName = un
		}
	}

	pBody, _ := h.db.Select("challenge_participants",
		fmt.Sprintf("session_id=eq.%s&select=*", challengeID), true)
	var participants []map[string]interface{}
	json.Unmarshal(pBody, &participants)
	userIDs := []string{}
	for _, p := range participants {
		uid := database.DBValue(p["user_id"])
		if uid != "" {
			userIDs = append(userIDs, uid)
		}
	}
	profilesMap := map[string]map[string]interface{}{}
	if len(userIDs) > 0 {
		profiles, _ := h.db.GetProfiles(userIDs)
		for _, p := range profiles {
			profilesMap[database.DBValue(p["id"])] = p
		}
	}
	type enrichedParticipant struct {
		UserID      string                 `json:"user_id"`
		XpBet       int                    `json:"xp_bet"`
		IndividualBet int                  `json:"individual_bet"`
		BetVersion  int                    `json:"bet_version"`
		BetAccepted bool                   `json:"bet_accepted"`
		Status      string                 `json:"status"`
		Score       *int                   `json:"score"`
		User        map[string]interface{} `json:"user"`
	}
	enriched := []enrichedParticipant{}
	for _, p := range participants {
		uid := database.DBValue(p["user_id"])
		ep := enrichedParticipant{
			UserID:        uid,
			XpBet:         database.DBInt(p["xp_bet"]),
			IndividualBet: database.DBInt(p["individual_bet"]),
			BetVersion:    database.DBInt(p["bet_version"]),
			BetAccepted:   database.DBBool(p["bet_accepted"]),
			Status:        database.DBValue(p["status"]),
			User:          profilesMap[uid],
		}
		if s, ok := p["score"].(float64); ok {
			v := int(s)
			ep.Score = &v
		}
		enriched = append(enriched, ep)
	}

	invBody, _ := h.db.Select("challenge_invitations",
		fmt.Sprintf("session_id=eq.%s&select=*", challengeID), true)
	var invitations []map[string]interface{}
	json.Unmarshal(invBody, &invitations)
	invUserIDs := []string{}
	for _, inv := range invitations {
		eid := database.DBValue(inv["invitee_id"])
		if eid != "" {
			invUserIDs = append(invUserIDs, eid)
		}
	}
	invProfilesMap := map[string]map[string]interface{}{}
	if len(invUserIDs) > 0 {
		invProfiles, _ := h.db.GetProfiles(invUserIDs)
		for _, p := range invProfiles {
			invProfilesMap[database.DBValue(p["id"])] = p
		}
	}
	type enrichedInvitation struct {
		ID       string                 `json:"id"`
		InviteeID string               `json:"invitee_id"`
		Status   string                 `json:"status"`
		Invitee  map[string]interface{} `json:"invitee"`
	}
	enrichedInv := []enrichedInvitation{}
	for _, inv := range invitations {
		eid := database.DBValue(inv["invitee_id"])
		ei := enrichedInvitation{
			ID:        database.DBValue(inv["id"]),
			InviteeID: eid,
			Status:    database.DBValue(inv["status"]),
			Invitee:   invProfilesMap[eid],
		}
		enrichedInv = append(enrichedInv, ei)
	}

	return c.JSON(fiber.Map{
		"session": map[string]interface{}{
			"id":              database.DBValue(session["id"]),
			"quiz_id":         quizID,
			"creator_id":      creatorID,
			"creator_name":    creatorName,
			"status":          database.DBValue(session["status"]),
			"total_xp_pool":   database.DBInt(session["total_xp_pool"]),
			"quiz_title":      quizTitle,
			"question_count":  questionCount,
		},
		"participants": enriched,
		"invitations":  enrichedInv,
	})
}

func (h *Handler) APIStartChallenge(c *fiber.Ctx) error {
	challengeID := c.Params("id")

	h.db.Update("challenge_sessions",
		fmt.Sprintf("id=eq.%s", challengeID),
		[]byte(fmt.Sprintf(`{"status":"playing","started_at":"%s"}`, time.Now().UTC().Format(time.RFC3339))), true)

	body, _ := h.db.Select("challenge_sessions",
		fmt.Sprintf("id=eq.%s&select=quiz_id", challengeID), true)
	var sessions []map[string]interface{}
	json.Unmarshal(body, &sessions)

	quizID := ""
	if len(sessions) > 0 {
		quizID, _ = sessions[0]["quiz_id"].(string)
	}

	return c.JSON(fiber.Map{"success": true, "quiz_id": quizID})
}

func (h *Handler) ChallengeCreate(c *fiber.Ctx) error {
	quizID := c.Params("id")
	user := h.getUserFromSession(c)
	if user == nil {
		return c.Redirect("/login")
	}

	quiz, err := h.db.GetQuiz(quizID)
	if err != nil {
		return c.Redirect("/explore")
	}

	title, _ := quiz["title"].(string)
	qCount := database.DBInt(quiz["question_count"])

	friendsBody, _ := h.db.Select("friendships",
		fmt.Sprintf("or=(requester_id.eq.%s,addressee_id.eq.%s)&status=eq.accepted&select=requester_id,addressee_id", user.ID, user.ID), true)
	var friendships []map[string]interface{}
	json.Unmarshal(friendsBody, &friendships)

	friendIDs := []string{}
	for _, f := range friendships {
		rid, _ := f["requester_id"].(string)
		aid, _ := f["addressee_id"].(string)
		if rid == user.ID && aid != "" {
			friendIDs = append(friendIDs, aid)
		} else if aid == user.ID && rid != "" {
			friendIDs = append(friendIDs, rid)
		}
	}

	friendsList := ""
	if len(friendIDs) > 0 {
		profiles, _ := h.db.GetProfiles(friendIDs)
		for _, p := range profiles {
			pName, _ := p["username"].(string)
			pNick, _ := p["nickname"].(string)
			pRank, _ := p["rank"].(string)
			pID, _ := p["id"].(string)
			displayName := pName
			if pNick != "" {
				displayName = pNick
			}
			friendsList += fmt.Sprintf(`<label class="friend-option"><input type="checkbox" name="invitees" value="%s"><span class="fo-avatar">%s</span><span class="fo-name">%s%s</span><span class="fo-rank rank-badge rank-%s">%s</span></label>`,
				pID, strings.ToUpper(string(pName[0])), displayName, challengeStatsBadgeMap(p), strings.ToLower(pRank), pRank)
		}
	}
	if friendsList == "" {
		friendsList = `<p class="text-muted text-center py-4">Ajoutez des amis pour les défier !</p>`
	}

	return renderWithContent(c, "Créer un défi", fmt.Sprintf(`
<div class="challenge-create max-w-2xl mx-auto">
    <a href="/quiz/%s" class="back-link">← Retour au quiz</a>
    <div class="cc-card">
        <div class="cc-header">⚔️ Créer un défi</div>
        <div class="cc-quiz-info">
            <div class="cc-quiz-title">%s</div>
            <div class="cc-quiz-meta">%d questions</div>
        </div>
        <form method="POST" action="/challenges/create/%s">
            <div class="cc-field">
                <label>Mise XP</label>
                <input type="number" name="xp_bet" value="100" min="10" max="10000" class="pe-input">
                <small>Votre solde: <strong>%d XP</strong> — Montant que vous misez (minimum 10)</small>
            </div>
            <div class="cc-field">
                <label>Inviter des amis</label>
                <div class="cc-friends">%s</div>
            </div>
            <button type="submit" class="btn-primary btn-lg w-full">⚔️ Créer le défi</button>
        </form>
    </div>
</div>`, quizID, title, qCount, quizID, user.XP, friendsList))
}

func (h *Handler) ChallengeCreatePost(c *fiber.Ctx) error {
	quizID := c.Params("id")
	user := h.getUserFromSession(c)
	if user == nil {
		return c.Redirect("/login")
	}

	xpBetStr := c.FormValue("xp_bet")
	xpBet := 100
	if x, err := strconv.Atoi(xpBetStr); err == nil && x >= 10 {
		xpBet = x
	}

	if user.XP < xpBet {
		return c.Redirect("/quiz/" + quizID + "?error=Vous n'avez pas assez d'XP pour miser")
	}

	quiz, _ := h.db.GetQuiz(quizID)
	quizTitle := "Quiz"
	if quiz != nil {
		quizTitle, _ = quiz["title"].(string)
	}

	sessionData := fmt.Sprintf(`{"quiz_id":"%s","creator_id":"%s","status":"waiting","total_xp_pool":%d,"invite_expires_at":"%s"}`,
		quizID, user.ID, xpBet, time.Now().UTC().Add(24*time.Hour).Format(time.RFC3339))
	sessionBody, err := h.db.Insert("challenge_sessions", []byte(sessionData), true)
	if err != nil {
		return c.Redirect("/quiz/" + quizID)
	}
	var sessions []map[string]interface{}
	json.Unmarshal(sessionBody, &sessions)

	sessionID := ""
	if len(sessions) > 0 {
		sessionID, _ = sessions[0]["id"].(string)
	}

	if sessionID == "" {
		return c.Redirect("/quiz/" + quizID)
	}

	participantData := fmt.Sprintf(`{"session_id":"%s","user_id":"%s","xp_bet":%d,"status":"accepted","individual_bet":%d,"bet_version":1,"bet_accepted":true}`,
		sessionID, user.ID, xpBet, xpBet)
	h.db.Insert("challenge_participants", []byte(participantData), true)

	inviterName := user.Username
	if user.Nickname != nil && *user.Nickname != "" {
		inviterName = *user.Nickname
	}

	invitees := []string{}
	formVals := c.Request().PostArgs().PeekMulti("invitees")
	if len(formVals) == 0 {
		formVals = c.Request().PostArgs().PeekMulti("invitees[]")
	}
	if len(formVals) == 0 {
		if mf, err := c.MultipartForm(); err == nil {
			formVals = make([][]byte, 0)
			for _, v := range mf.Value["invitees"] {
				formVals = append(formVals, []byte(v))
			}
		}
	}
	for _, v := range formVals {
		invitees = append(invitees, string(v))
	}
	if len(invitees) == 0 {
		log.Printf("[ChallengeCreatePost] raw body: %s", string(c.Request().Body()))
	}
	log.Printf("[ChallengeCreatePost] invitees: %v (xp_bet=%d)", invitees, xpBet)
	for _, inviteeID := range invitees {
		inviteeID = strings.TrimSpace(inviteeID)
		if inviteeID == "" || inviteeID == user.ID {
			continue
		}

		invData := fmt.Sprintf(`{"session_id":"%s","inviter_id":"%s","invitee_id":"%s","status":"pending","expires_at":"%s"}`, sessionID, user.ID, inviteeID, time.Now().UTC().Add(24*time.Hour).Format(time.RFC3339))
		if _, err := h.db.Insert("challenge_invitations", []byte(invData), true); err != nil {
			log.Printf("[ChallengeCreatePost] insert invitation error: %v", err)
		}

		createNotification(h.db, inviteeID, "challenge_invitation",
			"⚔️ Défi reçu !",
			fmt.Sprintf("%s vous a défié sur « %s » ! Mise : %d XP", inviterName, quizTitle, xpBet),
			fmt.Sprintf(`{"session_id":"%s","inviter_id":"%s"}`, sessionID, user.ID))

		h.sendSystemMessage(user.ID, inviteeID,
			fmt.Sprintf("⚔️ %s vous a défié sur « %s » ! Mise : %d XP. Acceptez dans l'onglet Défis.", inviterName, quizTitle, xpBet))
	}
	return c.Redirect("/challenges/" + sessionID)
}

func (h *Handler) sendSystemMessage(senderID, receiverID, content string) {
	smallerID, biggerID := senderID, receiverID
	if smallerID > biggerID {
		smallerID, biggerID = biggerID, smallerID
	}

	checkBody, err := h.db.Select("conversations",
		fmt.Sprintf("user1_id=eq.%s&user2_id=eq.%s&select=id", smallerID, biggerID), true)
	convID := ""
	if err == nil {
		var existing []map[string]interface{}
		json.Unmarshal(checkBody, &existing)
		if len(existing) > 0 {
			convID, _ = existing[0]["id"].(string)
		}
	}

	if convID == "" {
		insertData := fmt.Sprintf(`{"user1_id":"%s","user2_id":"%s"}`, smallerID, biggerID)
		insertBody, err := h.db.Insert("conversations", []byte(insertData), true)
		if err == nil {
			var created []map[string]interface{}
			json.Unmarshal(insertBody, &created)
			if len(created) > 0 {
				convID, _ = created[0]["id"].(string)
			}
		}
	}

	if convID == "" {
		return
	}

	msgData, _ := json.Marshal(map[string]interface{}{
		"conversation_id": convID,
		"sender_id":       senderID,
		"content":         content,
	})
	h.db.Insert("messages", msgData, true)

	h.db.Update("conversations",
		fmt.Sprintf("id=eq.%s", convID),
		[]byte(fmt.Sprintf(`{"last_message_at":"%s"}`, time.Now().UTC().Format(time.RFC3339))),
		true)
}

func parseInt(s string) (int, error) {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
}

func (h *Handler) ProfileView(c *fiber.Ctx) error {
	username := c.Params("username")

	currentUser := h.getUserFromSession(c)

	profilesBody, _ := h.db.Select("user_profiles",
		fmt.Sprintf("username=eq.%s&select=*", username), false)
	var profiles []map[string]interface{}
	json.Unmarshal(profilesBody, &profiles)
	if len(profiles) == 0 {
		return c.Redirect("/dashboard")
	}
	profile := profiles[0]
	profileID, _ := profile["id"].(string)
	pUsername, _ := profile["username"].(string)
	pNickname := database.DBValue(profile["nickname"])
	pRank, _ := profile["rank"].(string)
	pLevel := database.DBInt(profile["level"])
	pAvatar := database.DBValue(profile["avatar_url"])
	pBio := database.DBValue(profile["bio"])
	pFav := database.DBValue(profile["favorite_anime"])
	pCountry := database.DBValue(profile["country"])

	isOwnProfile := currentUser != nil && currentUser.ID == profileID

	// Statut amitié
	friendBtnHTML := ""
	if !isOwnProfile && currentUser != nil {
		supabaseURL := os.Getenv("SUPABASE_URL")
		anonKey := os.Getenv("SUPABASE_ANON_KEY")
		fsURL := fmt.Sprintf("%s/rest/v1/friendships?or=(and(requester_id.eq.%s,addressee_id.eq.%s),and(requester_id.eq.%s,addressee_id.eq.%s))&select=id,status,requester_id", supabaseURL, currentUser.ID, profileID, profileID, currentUser.ID)
		fsReq, _ := http.NewRequest("GET", fsURL, nil)
		fsReq.Header.Set("apikey", anonKey)
		fsResp, _ := http.DefaultClient.Do(fsReq)
		fsData := []map[string]interface{}{}
		if fsResp != nil {
			defer fsResp.Body.Close()
			body, _ := io.ReadAll(fsResp.Body)
			json.Unmarshal(body, &fsData)
		}
		if len(fsData) > 0 {
			f := fsData[0]
			fStatus, _ := f["status"].(string)
			fID, _ := f["id"].(string)
			rID, _ := f["requester_id"].(string)
			isReq := rID == currentUser.ID
			if fStatus == "accepted" {
				friendBtnHTML = fmt.Sprintf(`<button class="btn-outline btn-sm friend-action-btn" data-fid="%s" data-action="remove" onclick="toggleFriend(this)">✅ Ami</button>`, fID)
			} else if fStatus == "pending" && isReq {
				friendBtnHTML = `<span class="badge-outline">Demande envoyée</span>`
			} else if fStatus == "pending" && !isReq {
				friendBtnHTML = fmt.Sprintf(`<div class="friend-action-group"><button class="btn-sm btn-success" onclick="acceptFriendReq('%s')">✓ Accepter</button><button class="btn-sm btn-outline ch-danger" onclick="rejectFriendReq('%s')">✕</button></div>`, fID, fID)
			}
		}
		if friendBtnHTML == "" {
			friendBtnHTML = fmt.Sprintf(`<button class="btn-primary btn-sm" onclick="sendFriendReq('%s',this)">+ Ajouter en ami</button>`, profileID)
		}
	}

	displayName := pUsername
	if pNickname != "" {
		displayName = pNickname
	}

	avatarHTML := fmt.Sprintf(`<div class="pv-avatar-initial">%s</div>`, strings.ToUpper(string(pUsername[0])))
	if pAvatar != "" {
		avatarHTML = fmt.Sprintf(`<img src="%s" class="pv-avatar-img" alt="">`, pAvatar)
	}

	statsBody, _ := h.db.RPC("get_user_stats", map[string]interface{}{"user_id": profileID})
	var stats map[string]interface{}
	json.Unmarshal(statsBody, &stats)

	badgesBody, _ := h.db.Select("user_badges",
		fmt.Sprintf("user_id=eq.%s&select=*,badge:badge_id(*)", profileID), false)
	var badges []map[string]interface{}
	json.Unmarshal(badgesBody, &badges)

	quizzesBody, _ := h.db.Select("quizzes",
		fmt.Sprintf("creator_id=eq.%s&order=created_at.desc&limit=20", profileID), false)
	var userQuizzes []map[string]interface{}
	json.Unmarshal(quizzesBody, &userQuizzes)

	totalPlayed := database.DBInt(stats["quizzes_played"])
	created := len(userQuizzes)
	accuracy := database.DBInt(stats["accuracy_rate"])
	bestScore := database.DBInt(stats["best_score_ever"])

	badgesHTML := ""
	for _, ub := range badges {
		badge, _ := ub["badge"].(map[string]interface{})
		bName := "Badge"
		if badge != nil {
			if n, ok := badge["name"].(string); ok {
				bName = n
			}
		}
		badgesHTML += fmt.Sprintf(`<div class="badge-item"><div class="badge-icon">🏅</div><div class="badge-name">%s</div></div>`, bName)
	}
	if badgesHTML == "" {
		badgesHTML = `<p class="text-muted text-center py-6">Aucun badge</p>`
	}

	quizzesHTML := ""
	for _, q := range userQuizzes {
		qTitle, _ := q["title"].(string)
		qID, _ := q["id"].(string)
		qCount := database.DBInt(q["question_count"])
		pCount := database.DBInt(q["play_count"])
		quizzesHTML += fmt.Sprintf(`<a href="/quiz/%s" class="qz-card"><div class="qz-title">%s</div><div class="qz-meta">%d questions &bull; %d plays</div></a>`, qID, qTitle, qCount, pCount)
	}
	if quizzesHTML == "" {
		quizzesHTML = `<p class="text-muted text-center py-6">Aucun quiz créé</p>`
	}

	extraInfo := ""
	if pBio != "" {
		extraInfo += `<p class="pv-bio">` + pBio + `</p>`
	}

	return renderWithContent(c, "Profil de "+pUsername, fmt.Sprintf(`
<div class="pv-container">
    <div class="pv-header">
        %s
        <div class="pv-info">
            <h1 class="pv-name">%s</h1>
            <div class="pv-meta">
                <span class="rank-badge rank-%s">%s</span>
                <span class="text-muted">Niveau %d</span>%s
            </div>
            %s
            <div class="pv-details">%s%s</div>
        </div>
        %s
    </div>

    <div class="d-stats">
        <div class="ds-card ds-brand"><div class="ds-icon">🎮</div><div class="ds-value">%d</div><div class="ds-label">Quiz joués</div></div>
        <div class="ds-card ds-accent"><div class="ds-icon">📝</div><div class="ds-value">%d</div><div class="ds-label">Quiz créés</div></div>
        <div class="ds-card ds-green"><div class="ds-icon">🎯</div><div class="ds-value">%d%%</div><div class="ds-label">Précision</div></div>
        <div class="ds-card ds-purple"><div class="ds-icon">🏆</div><div class="ds-value">%d</div><div class="ds-label">Meilleur score</div></div>
    </div>

    <div class="pv-tabs">
        <button class="pv-tab active" onclick="switchProfileTab('badges', this)">🏅 Badges (%d)</button>
        <button class="pv-tab" onclick="switchProfileTab('quizzes', this)">📚 Quiz (%d)</button>
    </div>

    <div id="pv-badges" class="pv-tab-content"><div class="badge-grid">%s</div></div>
    <div id="pv-quizzes" class="pv-tab-content" style="display:none"><div class="qz-grid">%s</div></div>
</div>
<script>
function switchProfileTab(name, btn) {
    document.querySelectorAll('.pv-tab').forEach(function(t) { t.classList.remove('active'); });
    document.querySelectorAll('.pv-tab-content').forEach(function(t) { t.style.display = 'none'; });
    btn.classList.add('active');
    document.getElementById('pv-' + name).style.display = 'block';
}
function sendFriendReq(userId, btn) {
    fetch('/api/friends/request', {method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({user_id:userId})})
        .then(function(r){return r.json()})
        .then(function(d){
            if(d.success&&btn){btn.outerHTML='<span class="badge-outline">Demande envoyée</span>';}
            else alert(d.error||'Erreur');
        });
}
function acceptFriendReq(fid) {
    fetch('/api/friends/accept',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({friendship_id:fid})})
        .then(function(r){return r.json()})
        .then(function(d){if(d.success)location.reload();});
}
function rejectFriendReq(fid) {
    fetch('/api/friends/reject',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({friendship_id:fid})})
        .then(function(r){return r.json()})
        .then(function(d){if(d.success)location.reload();});
}
function toggleFriend(btn) {
    var fid=btn.getAttribute('data-fid');
    var action=btn.getAttribute('data-action');
    if(action==='remove'){
        if(!confirm('Supprimer cet ami ?'))return;
        fetch('/api/friends/remove',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({friendship_id:fid})})
            .then(function(){location.reload();});
    }
}
</script>`,
		avatarHTML, displayName,
		strings.ToLower(pRank), pRank, pLevel,
		challengeStatsBadgeMap(profile),
		extraInfo,
		func() string { if pCountry != "" { return "📍 " + pCountry + " " }; return "" }(),
		func() string { if pFav != "" { return "❤️ " + pFav }; return "" }(),
		friendBtnHTML,
		totalPlayed, created, accuracy, bestScore,
		len(badges), len(userQuizzes),
		badgesHTML, quizzesHTML,
	))
}

func (h *Handler) Events(c *fiber.Ctx) error {
	activeEvents, _ := h.db.GetQuizzes("quiz_type=eq.official&is_visible=eq.true&starts_at=lte.now&ends_at=gte.now&order=starts_at.desc")
	upcomingEvents, _ := h.db.GetQuizzes("quiz_type=eq.official&is_visible=eq.true&starts_at=gt.now&order=starts_at.asc")

	activeHTML := ""
	for _, e := range activeEvents {
		activeHTML += h.renderEventCardLarge(e)
	}
	upcomingHTML := ""
	for _, e := range upcomingEvents {
		upcomingHTML += h.renderEventCardLarge(e)
	}

	if activeHTML == "" {
		activeHTML = `<p class="text-muted text-center py-8">Aucun événement en cours</p>`
	}
	if upcomingHTML == "" {
		upcomingHTML = `<p class="text-muted text-center py-8">Aucun événement à venir</p>`
	}

	return renderWithContent(c, "Événements", fmt.Sprintf(`
<div class="events-page max-w-4xl mx-auto">
    <h1 class="page-title">🎪 Événements</h1>
    <section class="mb-8">
        <h2 class="d-section-title"><span style="color:#22c55e">🔴</span> EN DIRECT</h2>
        <div class="space-y-4">%s</div>
    </section>
    <section>
        <h2 class="d-section-title">📅 À VENIR</h2>
        <div class="space-y-4">%s</div>
    </section>
</div>`, activeHTML, upcomingHTML))
}

func (h *Handler) Badges(c *fiber.Ctx) error {
	user := h.getUserFromSession(c)
	if user == nil {
		return c.Redirect("/login")
	}

	statsBody, _ := h.db.Select("user_stats",
		fmt.Sprintf("user_id=eq.%s&select=quizzes_played,quizzes_created,accuracy_rate,total_correct_answers,total_answers,best_score_ever", user.ID), true)
	var stats []map[string]interface{}
	json.Unmarshal(statsBody, &stats)
	quizzesPlayed := 0
	quizzesCreated := 0
	accuracy := 0.0
	perfectCount := 0
	if len(stats) > 0 {
		quizzesPlayed = database.DBInt(stats[0]["quizzes_played"])
		quizzesCreated = database.DBInt(stats[0]["quizzes_created"])
		accuracy, _ = stats[0]["accuracy_rate"].(float64)
	}

	perfectBody, _ := h.db.Select("game_sessions",
		fmt.Sprintf("user_id=eq.%s&is_perfect=eq.true&select=id", user.ID), true)
	var perfects []map[string]interface{}
	json.Unmarshal(perfectBody, &perfects)
	perfectCount = len(perfects)

	isMonthlyTop10 := false
	isMonthlyChampion := false
	rpcBody, err := h.db.RPC("rpc_get_user_monthly_rank", map[string]interface{}{
		"target_user_id": user.ID,
	})
	if err == nil {
		var monthlyRank []map[string]interface{}
		json.Unmarshal(rpcBody, &monthlyRank)
		if len(monthlyRank) > 0 {
			rank := database.DBInt(monthlyRank[0]["rank"])
			if rank > 0 && rank <= 10 {
				isMonthlyTop10 = true
			}
			if rank == 1 {
				isMonthlyChampion = true
			}
		}
	}

	badgesBody, _ := h.db.Select("badges", "order=condition_value.asc", false)
	var allBadges []map[string]interface{}
	json.Unmarshal(badgesBody, &allBadges)

	if len(allBadges) == 0 {
		h.seedBadges()
		badgesBody, _ = h.db.Select("badges", "order=condition_value.asc", true)
		json.Unmarshal(badgesBody, &allBadges)
	}

	ownedBody, _ := h.db.Select("user_badges",
		fmt.Sprintf("user_id=eq.%s&select=badge_id,earned_at", user.ID), true)
	var ownedRows []map[string]interface{}
	json.Unmarshal(ownedBody, &ownedRows)
	ownedMap := map[string]time.Time{}
	for _, ub := range ownedRows {
		earnedStr := database.DBValue(ub["earned_at"])
		t, _ := time.Parse(time.RFC3339, earnedStr)
		ownedMap[database.DBValue(ub["badge_id"])] = t
	}

	slugIcons := map[string]string{
		"first_quiz": "🎮", "quiz_10": "🎯", "quiz_100": "🏆", "quiz_500": "👑", "quiz_1000": "💎",
		"perfect_quiz": "⭐", "perfect_5": "🌟", "accuracy_90": "🎯",
		"first_creation": "✍️", "creator_10": "🎨", "popular_creator": "🔥", "elite_creator": "💫",
		"top10_monthly": "🏅", "monthly_champion": "🥇",
		"speed_demon": "⚡", "streak_master": "🔥", "night_owl": "🦉", "weekend_warrior": "⚔️",
	}

	conditionLabels := map[string]string{
		"quizzes_played":  "Quiz complétés",
		"quizzes_created": "Quiz créés",
		"perfect_quiz":    "Scores parfaits",
		"accuracy_rate":   "Précision",
		"popular_quiz":    "Joueurs sur un quiz",
		"monthly_top10":   "Top 10 mensuel",
		"monthly_champion": "Champion du mois",
		"speed_answer":    "Vitesse de réponse",
		"streak":          "Série de bonnes réponses",
		"night_play":      "Jouer la nuit",
		"weekend_play":    "Quiz le week-end",
	}

	conditionDetail := func(slug, cType string, cValue int) string {
		switch cType {
		case "quizzes_played":
			return fmt.Sprintf("Complète %d quiz au total", cValue)
		case "quizzes_created":
			return fmt.Sprintf("Crée %d quiz", cValue)
		case "perfect_quiz":
			return fmt.Sprintf("Obtiens %d score(s) parfait(s)", cValue)
		case "accuracy_rate":
			return fmt.Sprintf("Atteins %d%% de précision (20+ quiz)", cValue)
		case "popular_quiz":
			return fmt.Sprintf("Un de tes quiz atteint %d joueurs", cValue)
		case "monthly_top10":
			return "Termine dans le top 10 du classement mensuel"
		case "monthly_champion":
			return "Termine 1er du classement mensuel"
		case "speed_answer":
			return "Réponds en moins de 2 secondes"
		case "streak":
			return fmt.Sprintf("Obtiens une série de %d bonnes réponses", cValue)
		case "night_play":
			return "Joue entre 0h et 5h du matin"
		case "weekend_play":
			return fmt.Sprintf("Joue %d quiz le week-end", cValue)
		}
		return ""
	}

	progressValue := func(cType string, cValue int) (int, int) {
		switch cType {
		case "quizzes_played":
			return quizzesPlayed, cValue
		case "quizzes_created":
			return quizzesCreated, cValue
		case "perfect_quiz":
			return perfectCount, cValue
		case "accuracy_rate":
			return int(accuracy), cValue
		case "monthly_top10":
			if isMonthlyTop10 { return 1, 1 }
			return 0, 1
		case "monthly_champion":
			if isMonthlyChampion { return 1, 1 }
			return 0, 1
		}
		return 0, cValue
	}

	earnedHTML := ""
	pendingHTML := ""

	for _, b := range allBadges {
		bID := database.DBValue(b["id"])
		slug := database.DBValue(b["slug"])
		bName := database.DBValue(b["name"])
		bDesc := database.DBValue(b["description"])
		isRare := database.DBBool(b["is_rare"])
		conditionType := database.DBValue(b["condition_type"])
		conditionValue := database.DBInt(b["condition_value"])

		icon := "🏅"
		if v, ok := slugIcons[slug]; ok {
			icon = v
		}

		rarityColor := "#94a3b8"
		rarityLabel := "Commun"
		if isRare {
			rarityColor = "#a855f7"
			rarityLabel = "Rare"
		}

		earned, hasEarned := ownedMap[bID]

		cLabel := conditionLabels[conditionType]
		cDetail := conditionDetail(slug, conditionType, conditionValue)

		if hasEarned {
			earnedDate := ""
			if !earned.IsZero() {
				earnedDate = earned.Format("02/01/2006")
			}
			earnedHTML += fmt.Sprintf(`
<div class="bcard bcard-earned" style="border-left:3px solid #22c55e">
<div class="bcard-top">
<div class="bcard-icon">%s</div>
<div class="bcard-info">
<div class="bcard-name">%s</div>
<div class="bcard-desc">%s</div>
<div class="bcard-rarity" style="color:%s">✦ %s</div>
</div>
<div class="bcard-status">✅<span style="font-size:.7rem;color:#64748b;display:block">Le %s</span></div>
</div>
</div>`, icon, bName, bDesc, rarityColor, rarityLabel, earnedDate)
		} else {
			cur, total := progressValue(conditionType, conditionValue)
			pct := 0
			if total > 0 {
				pct = (cur * 100) / total
			}
			if pct > 100 { pct = 100 }

			pendingHTML += fmt.Sprintf(`
<div class="bcard bcard-pending">
<div class="bcard-top">
<div class="bcard-icon bcard-icon-locked">%s</div>
<div class="bcard-info">
<div class="bcard-name">%s</div>
<div class="bcard-desc">%s</div>
<div class="bcard-condition">
<span class="bcard-clabel">%s</span>
<span class="bcard-ctext">%s</span>
</div>
</div>
<div class="bcard-rarity" style="color:%s">✦ %s</div>
</div>
<div class="bcard-progress">
<div class="bcard-bar"><div class="bcard-bar-fill" style="width:%d%%;background:%s"></div></div>
<div class="bcard-bar-label">%d / %d</div>
</div>
</div>`, icon, bName, bDesc, cLabel, cDetail, rarityColor, rarityLabel, pct, rarityColor, cur, total)
		}
	}

	if earnedHTML == "" {
		earnedHTML = `<div class="bcard-empty">Aucun badge débloqué pour l'instant. Continue à jouer !</div>`
	}
	if pendingHTML == "" {
		pendingHTML = `<div class="bcard-empty">Félicitations, tu as tous les badges ! 🎉</div>`
	}

	earnedCount := len(ownedMap)
	totalCount := len(allBadges)

	content := `
<div class="badges-page">
<div class="badges-header">
<h1>🏅 Badges</h1>
<div class="badges-stat"><span class="badges-stat-num">` + fmt.Sprintf("%d", earnedCount) + `</span> / ` + fmt.Sprintf("%d", totalCount) + ` débloqués</div>
</div>

<div class="badges-section">
<h2 class="badges-section-title">✅ Badges débloqués (` + fmt.Sprintf("%d", earnedCount) + `)</h2>
<div class="badges-grid">
` + earnedHTML + `
</div>
</div>

<div class="badges-section">
<h2 class="badges-section-title">🎯 Badges à débloquer (` + fmt.Sprintf("%d", totalCount-earnedCount) + `)</h2>
<div class="badges-grid">
` + pendingHTML + `
</div>
</div>
</div>

<style>
.badges-page{max-width:800px;margin:0 auto;padding:0 16px}
.badges-header{display:flex;align-items:center;justify-content:space-between;margin-bottom:24px}
.badges-header h1{margin:0}
.badges-stat{font-size:.9rem;color:#94a3b8}
.badges-stat-num{color:#6366f1;font-weight:700;font-size:1.1rem}
.badges-section{margin-bottom:28px}
.badges-section-title{font-size:1rem;color:#94a3b8;margin-bottom:12px;padding-bottom:8px;border-bottom:1px solid #2d2d44}
.badges-grid{display:flex;flex-direction:column;gap:8px}

.bcard{background:#16213e;border:1px solid #2d2d44;border-radius:10px;padding:14px 16px;transition:all .2s}
.bcard:hover{border-color:#6366f1}
.bcard-earned{background:rgba(34,197,94,.03)}
.bcard-top{display:flex;align-items:flex-start;gap:12px}
.bcard-icon{font-size:1.6rem;width:40px;height:40px;display:flex;align-items:center;justify-content:center;background:#1e293b;border-radius:8px;flex-shrink:0}
.bcard-icon-locked{opacity:.4;filter:grayscale(1)}
.bcard-info{flex:1;min-width:0}
.bcard-name{font-weight:600;font-size:.95rem}
.bcard-desc{font-size:.8rem;color:#94a3b8;margin-top:2px}
.bcard-rarity{font-size:.7rem;margin-top:3px;font-weight:600;text-transform:uppercase}
.bcard-status{text-align:center;flex-shrink:0}
.bcard-condition{display:flex;gap:6px;align-items:center;margin-top:6px;flex-wrap:wrap}
.bcard-clabel{font-size:.7rem;background:#1e293b;color:#94a3b8;padding:2px 8px;border-radius:4px;font-weight:600}
.bcard-ctext{font-size:.75rem;color:#64748b}
.bcard-progress{margin-top:10px}
.bcard-bar{background:#1e293b;border-radius:4px;height:6px;overflow:hidden}
.bcard-bar-fill{height:100%;border-radius:4px;transition:width .5s}
.bcard-bar-label{font-size:.7rem;color:#64748b;margin-top:3px;text-align:right}
.bcard-empty{text-align:center;padding:30px;color:#64748b;font-size:.9rem}
</style>`

	return renderWithContent(c, "Badges", content)
}

func (h *Handler) seedBadges() {
	badges := []map[string]interface{}{
		{"slug": "first_quiz", "name": "Premier Pas", "description": "Complète ton premier quiz", "condition_type": "quizzes_played", "condition_value": 1, "is_rare": false},
		{"slug": "quiz_10", "name": "Débutant", "description": "Complète 10 quiz", "condition_type": "quizzes_played", "condition_value": 10, "is_rare": false},
		{"slug": "quiz_100", "name": "Vétéran", "description": "Complète 100 quiz", "condition_type": "quizzes_played", "condition_value": 100, "is_rare": false},
		{"slug": "quiz_500", "name": "Légende Vivante", "description": "Complète 500 quiz", "condition_type": "quizzes_played", "condition_value": 500, "is_rare": true},
		{"slug": "quiz_1000", "name": "Immortel", "description": "Complète 1000 quiz", "condition_type": "quizzes_played", "condition_value": 1000, "is_rare": true},
		{"slug": "perfect_quiz", "name": "Perfection", "description": "Obtiens un score parfait (100%)", "condition_type": "perfect_quiz", "condition_value": 1, "is_rare": false},
		{"slug": "perfect_5", "name": "Maître de la Perfection", "description": "Obtiens 5 scores parfaits", "condition_type": "perfect_quiz", "condition_value": 5, "is_rare": true},
		{"slug": "accuracy_90", "name": "Précision Chirurgicale", "description": "90% de précision sur 20+ quiz", "condition_type": "accuracy_rate", "condition_value": 90, "is_rare": true},
		{"slug": "first_creation", "name": "Créateur", "description": "Crée ton premier quiz", "condition_type": "quizzes_created", "condition_value": 1, "is_rare": false},
		{"slug": "creator_10", "name": "Producteur", "description": "Crée 10 quiz", "condition_type": "quizzes_created", "condition_value": 10, "is_rare": false},
		{"slug": "popular_creator", "name": "Star Montante", "description": "Un quiz avec 100+ joueurs", "condition_type": "popular_quiz", "condition_value": 100, "is_rare": true},
		{"slug": "elite_creator", "name": "Superstar", "description": "Un quiz avec 1000+ joueurs", "condition_type": "popular_quiz", "condition_value": 1000, "is_rare": true},
		{"slug": "top10_monthly", "name": "Top 10", "description": "Termine dans le top 10 mensuel", "condition_type": "monthly_top10", "condition_value": 10, "is_rare": true},
		{"slug": "monthly_champion", "name": "Champion du Mois", "description": "Termine 1er du classement mensuel", "condition_type": "monthly_champion", "condition_value": 1, "is_rare": true},
		{"slug": "speed_demon", "name": "Démon de Vitesse", "description": "Réponds en moins de 2 secondes", "condition_type": "speed_answer", "condition_value": 2, "is_rare": false},
		{"slug": "streak_master", "name": "Maître des Séries", "description": "Série de 10 bonnes réponses", "condition_type": "streak", "condition_value": 10, "is_rare": true},
		{"slug": "night_owl", "name": "Chouette Nocturne", "description": "Joue à 3h du matin", "condition_type": "night_play", "condition_value": 1, "is_rare": false},
		{"slug": "weekend_warrior", "name": "Guerrier du Week-end", "description": "Joue 10 quiz le week-end", "condition_type": "weekend_play", "condition_value": 10, "is_rare": false},
	}
	for _, b := range badges {
		data, _ := json.Marshal(b)
		h.db.Insert("badges", data, true)
	}
}

func (h *Handler) Collections(c *fiber.Ctx) error {
	user := h.getUserFromSession(c)
	if user == nil {
		return c.Redirect("/login")
	}

	collectionsBody, _ := h.db.RPC("get_user_collections", map[string]interface{}{"user_id": user.ID})
	var collections []map[string]interface{}
	json.Unmarshal(collectionsBody, &collections)

	colHTML := ""
	for _, col := range collections {
		series, _ := col["series"].(string)
		completed := database.DBInt(col["completed_quizzes"])
		total := database.DBInt(col["total_quizzes"])
		progress := database.DBInt(col["progress_percent"])
		colHTML += fmt.Sprintf(`
<div class="col-card">
    <div class="col-header"><div class="col-series">📚 %s</div><div class="col-pct">%d/%d</div></div>
    <div class="col-bar"><div class="col-bar-fill" style="width:%d%%"></div></div>
    <div class="col-meta"><span>%d%% complété</span></div>
</div>`, series, completed, total, progress, progress)
	}
	if colHTML == "" {
		colHTML = `<p class="text-muted text-center py-8">Aucune collection commencée</p>`
	}

	return renderWithContent(c, "Collections", fmt.Sprintf(`
<div class="collections-page max-w-4xl mx-auto">
    <h1 class="page-title">📚 Collections</h1>
    <div class="col-grid">%s</div>
</div>`, colHTML))
}

func (h *Handler) Series(c *fiber.Ctx) error {
	seriesName := c.Params("series")
	quizzes, _ := h.db.GetQuizzes(fmt.Sprintf("series=ilike.*%s*&is_visible=eq.true&order=play_count.desc", seriesName))

	quizHTML := ""
	for _, q := range quizzes {
		quizHTML += h.renderQuizCard(q)
	}
	if quizHTML == "" {
		quizHTML = `<p class="text-muted text-center py-8">Aucun quiz pour cette série</p>`
	}

	return renderWithContent(c, seriesName, fmt.Sprintf(`
<div class="series-page max-w-4xl mx-auto">
    <a href="/explore" class="back-link">← Retour</a>
    <h1 class="page-title">📺 %s</h1>
    <div class="quiz-grid">%s</div>
</div>`, seriesName, quizHTML))
}

// ===== NEW API ENDPOINTS =====

func (h *Handler) APIGetSentRequests(c *fiber.Ctx) error {
	userID := c.Query("user_id")

	body, err := h.db.Select("friendships",
		fmt.Sprintf("requester_id=eq.%s&status=eq.pending", userID), true)
	if err != nil {
		return c.JSON(fiber.Map{"requests": []interface{}{}, "error": err.Error()})
	}
	var friendships []map[string]interface{}
	json.Unmarshal(body, &friendships)

	requests := []map[string]interface{}{}
	for _, f := range friendships {
		addresseeID, _ := f["addressee_id"].(string)
		friendshipID, _ := f["id"].(string)

		pBody, err := h.db.Select("user_profiles",
			fmt.Sprintf("id=eq.%s&select=id,username,nickname,avatar_url,rank,level", addresseeID), false)
		if err != nil {
			continue
		}
		var profiles []map[string]interface{}
		json.Unmarshal(pBody, &profiles)
		if len(profiles) > 0 {
			p := profiles[0]
			p["friendship_id"] = friendshipID
			requests = append(requests, p)
		}
	}

	return c.JSON(fiber.Map{"requests": requests})
}

func (h *Handler) APIGetOrCreateConversation(c *fiber.Ctx) error {
	type Request struct {
		FriendID string `json:"friend_id"`
	}
	var req Request
	c.BodyParser(&req)

	sess, _ := h.store.Get(c)
	userID := sess.Get("user_id").(string)

	smallerID, biggerID := userID, req.FriendID
	if smallerID > biggerID {
		smallerID, biggerID = biggerID, smallerID
	}

	checkBody, err := h.db.Select("conversations",
		fmt.Sprintf("user1_id=eq.%s&user2_id=eq.%s&select=id", smallerID, biggerID), true)
	if err == nil {
		var existing []map[string]interface{}
		json.Unmarshal(checkBody, &existing)
		if len(existing) > 0 {
			if id, ok := existing[0]["id"].(string); ok {
				return c.JSON(fiber.Map{"conversation_id": id})
			}
		}
	}

	insertData := fmt.Sprintf(`{"user1_id":"%s","user2_id":"%s"}`, smallerID, biggerID)
	insertBody, err := h.db.Insert("conversations", []byte(insertData), true)
	if err != nil {
		return c.JSON(fiber.Map{"conversation_id": "", "error": err.Error()})
	}
	var created []map[string]interface{}
	json.Unmarshal(insertBody, &created)
	if len(created) > 0 {
		if id, ok := created[0]["id"].(string); ok {
			return c.JSON(fiber.Map{"conversation_id": id})
		}
	}

	return c.JSON(fiber.Map{"conversation_id": ""})
}

func (h *Handler) APIGetChallengeInvitations(c *fiber.Ctx) error {
	userID := c.Query("user_id")

	body, err := h.db.Select("challenge_invitations",
		fmt.Sprintf("invitee_id=eq.%s&status=eq.pending", userID), true)
	if err != nil {
		return c.JSON(fiber.Map{"invitations": []interface{}{}, "error": err.Error()})
	}
	var invitations []map[string]interface{}
	json.Unmarshal(body, &invitations)

	result := []map[string]interface{}{}
	for _, inv := range invitations {
		inviterID, _ := inv["inviter_id"].(string)
		sessionID, _ := inv["session_id"].(string)

		quizTitle := "Quiz"
		sBody, err := h.db.Select("challenge_sessions",
			fmt.Sprintf("id=eq.%s&select=quiz:quiz_id(title)", sessionID), true)
		if err == nil {
			var sessions []map[string]interface{}
			json.Unmarshal(sBody, &sessions)
			if len(sessions) > 0 {
				if quiz, ok := sessions[0]["quiz"].(map[string]interface{}); ok {
					if t, ok := quiz["title"].(string); ok {
						quizTitle = t
					}
				}
			}
		}

		inviterName := "Quelqu'un"
		pBody, err := h.db.Select("user_profiles",
			fmt.Sprintf("id=eq.%s&select=username,nickname", inviterID), false)
		if err == nil {
			var profiles []map[string]interface{}
			json.Unmarshal(pBody, &profiles)
			if len(profiles) > 0 {
				if nn, ok := profiles[0]["nickname"].(string); ok && nn != "" {
					inviterName = nn
				} else if un, ok := profiles[0]["username"].(string); ok {
					inviterName = un
				}
			}
		}

		result = append(result, map[string]interface{}{
			"id":               inv["id"],
			"session_id":       sessionID,
			"quiz_title":       quizTitle,
			"inviter_username": inviterName,
			"inviter_nickname": inviterName,
			"created_at":       inv["created_at"],
		})
	}

	return c.JSON(fiber.Map{"invitations": result})
}

func (h *Handler) APIAcceptChallengeInvitation(c *fiber.Ctx) error {
	type Request struct {
		InvitationID string `json:"invitation_id"`
	}
	var req Request
	c.BodyParser(&req)

	sess, _ := h.store.Get(c)
	userID := sess.Get("user_id").(string)

	getBody, err := h.db.Select("challenge_invitations",
		fmt.Sprintf("id=eq.%s&select=session_id", req.InvitationID), true)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Invitation introuvable"})
	}
	var invs []map[string]interface{}
	json.Unmarshal(getBody, &invs)
	if len(invs) == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Invitation introuvable"})
	}
	sessionID, _ := invs[0]["session_id"].(string)

	if _, err := h.db.Update("challenge_invitations", fmt.Sprintf("id=eq.%s", req.InvitationID),
		[]byte(`{"status":"accepted"}`), true); err != nil {
		log.Printf("[APIAcceptChallengeInvitation] update invitation error: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "Erreur mise à jour invitation"})
	}

	inviterID := ""
	invBody, _ := h.db.Select("challenge_invitations",
		fmt.Sprintf("id=eq.%s&select=inviter_id", req.InvitationID), true)
	var invRows []map[string]interface{}
	json.Unmarshal(invBody, &invRows)
	if len(invRows) > 0 {
		inviterID, _ = invRows[0]["inviter_id"].(string)
	}

	totalPool := 100
	sBody, _ := h.db.Select("challenge_sessions",
		fmt.Sprintf("id=eq.%s&select=total_xp_pool", sessionID), true)
	var sRows []map[string]interface{}
	json.Unmarshal(sBody, &sRows)
	if len(sRows) > 0 {
		if p, ok := sRows[0]["total_xp_pool"].(float64); ok {
			totalPool = int(p)
		}
	}
	inviteeBet := totalPool
	if _, err := h.db.Update("challenge_sessions", fmt.Sprintf("id=eq.%s", sessionID),
		[]byte(fmt.Sprintf(`{"total_xp_pool":%d}`, totalPool+inviteeBet)), true); err != nil {
		log.Printf("[APIAcceptChallengeInvitation] update session error: %v", err)
	}

	partData := fmt.Sprintf(`{"session_id":"%s","user_id":"%s","xp_bet":%d,"status":"accepted","individual_bet":%d,"bet_version":1,"bet_accepted":false}`,
		sessionID, userID, inviteeBet, inviteeBet)
	respBody, insertErr := h.db.Insert("challenge_participants", []byte(partData), true)
	log.Printf("[APIAcceptChallengeInvitation] insert participant resp: %s err: %v", string(respBody), insertErr)
	if insertErr != nil {
		return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("Erreur ajout participant: %v", insertErr)})
	}
	if len(respBody) > 0 && !bytes.Contains(respBody, []byte("id")) {
		log.Printf("[APIAcceptChallengeInvitation] insert participant unexpected response: %s", string(respBody))
	}

	acceptorName := "Quelqu'un"
	aBody, _ := h.db.Select("user_profiles",
		fmt.Sprintf("id=eq.%s&select=username,nickname", userID), false)
	var aProfiles []map[string]interface{}
	json.Unmarshal(aBody, &aProfiles)
	if len(aProfiles) > 0 {
		if nn, ok := aProfiles[0]["nickname"].(string); ok && nn != "" {
			acceptorName = nn
		} else if un, ok := aProfiles[0]["username"].(string); ok {
			acceptorName = un
		}
	}

	if inviterID != "" {
		createNotification(h.db, inviterID, "challenge_invitation",
			"⚔️ Défi accepté !",
			fmt.Sprintf("%s a accepté votre défi !", acceptorName))
		h.sendSystemMessage(userID, inviterID,
			fmt.Sprintf("⚔️ %s a accepté votre défi ! La phase de négociation peut commencer.", acceptorName))
	}

	h.checkChallengeConsensus(sessionID)

	return c.JSON(fiber.Map{"success": true})
}

func (h *Handler) APIRefuseChallengeInvitation(c *fiber.Ctx) error {
	type Request struct {
		InvitationID string `json:"invitation_id"`
	}
	var req Request
	c.BodyParser(&req)

	inviterID := ""
	invBody, _ := h.db.Select("challenge_invitations",
		fmt.Sprintf("id=eq.%s&select=inviter_id,session_id", req.InvitationID), true)
	var invRows []map[string]interface{}
	json.Unmarshal(invBody, &invRows)
	sessionID := ""
	if len(invRows) > 0 {
		inviterID, _ = invRows[0]["inviter_id"].(string)
		sessionID, _ = invRows[0]["session_id"].(string)
	}

	_, err := h.db.Update("challenge_invitations", fmt.Sprintf("id=eq.%s", req.InvitationID),
		[]byte(`{"status":"refused"}`), true)
	if err != nil {
		return c.JSON(fiber.Map{"error": err.Error()})
	}

	refuserName := "Quelqu'un"
	rBody, _ := h.db.Select("user_profiles",
		fmt.Sprintf("id=eq.%s&select=username,nickname", c.Locals("user").(*UserProfile).ID), false)
	var rProfiles []map[string]interface{}
	json.Unmarshal(rBody, &rProfiles)
	if len(rProfiles) > 0 {
		if nn, ok := rProfiles[0]["nickname"].(string); ok && nn != "" {
			refuserName = nn
		} else if un, ok := rProfiles[0]["username"].(string); ok {
			refuserName = un
		}
	}

	if inviterID != "" {
		createNotification(h.db, inviterID, "challenge_invitation",
			"❌ Défi refusé",
			fmt.Sprintf("%s a refusé votre défi.", refuserName))
	}

	_ = sessionID
	return c.JSON(fiber.Map{"success": true})
}

func (h *Handler) APIGetMyChallenges(c *fiber.Ctx) error {
	userID := c.Query("user_id")

	partBody, err := h.db.Select("challenge_participants",
		fmt.Sprintf("user_id=eq.%s&select=session_id", userID), true)
	if err != nil {
		return c.JSON(fiber.Map{"challenges": []interface{}{}, "error": err.Error()})
	}
	var participations []map[string]interface{}
	json.Unmarshal(partBody, &participations)

	createdBody, err := h.db.Select("challenge_sessions",
		fmt.Sprintf("creator_id=eq.%s&select=id", userID), true)
	if err == nil {
		var created []map[string]interface{}
		json.Unmarshal(createdBody, &created)
		for _, cr := range created {
			if id, ok := cr["id"].(string); ok {
				participations = append(participations, map[string]interface{}{"session_id": id})
			}
		}
	}

	sessionIDs := map[string]bool{}
	for _, p := range participations {
		if id, ok := p["session_id"].(string); ok {
			sessionIDs[id] = true
		}
	}

	ids := []string{}
	for id := range sessionIDs {
		ids = append(ids, id)
	}

	result := []map[string]interface{}{}
	if len(ids) > 0 {
		idFilter := ""
		for i, id := range ids {
			if i > 0 {
				idFilter += ","
			}
			idFilter += "id.eq." + id
		}
		sessBody, err := h.db.Select("challenge_sessions",
			fmt.Sprintf("or=(%s)&select=*,quiz:quiz_id(title),participants:challenge_participants(*)", idFilter), true)
		if err == nil {
			var challenges []map[string]interface{}
			json.Unmarshal(sessBody, &challenges)
			for _, ch := range challenges {
				quizTitle := "Quiz"
				if quiz, ok := ch["quiz"].(map[string]interface{}); ok {
					if t, ok := quiz["title"].(string); ok {
						quizTitle = t
					}
				}
				st, _ := ch["status"].(string)
				if st == "cancelled" {
					continue
				}
				participants, _ := ch["participants"].([]interface{})
				result = append(result, map[string]interface{}{
					"id":                ch["id"],
					"quiz_title":        quizTitle,
					"status":            ch["status"],
					"total_xp_pool":     database.DBInt(ch["total_xp_pool"]),
					"participant_count": len(participants),
				})
			}
		}
	}

	return c.JSON(fiber.Map{"challenges": result})
}

func (h *Handler) APIDeleteConversation(c *fiber.Ctx) error {
	type Request struct {
		ConversationID string `json:"conversation_id"`
	}
	var req Request
	c.BodyParser(&req)

	sess, _ := h.store.Get(c)
	userID := sess.Get("user_id").(string)

	h.db.Delete("messages", fmt.Sprintf("conversation_id=eq.%s&sender_id=eq.%s", req.ConversationID, userID), true)

	return c.JSON(fiber.Map{"success": true})
}

func (h *Handler) APIReportUser(c *fiber.Ctx) error {
	type Request struct {
		UserID      string `json:"user_id"`
		Reason      string `json:"reason"`
		Description string `json:"description"`
	}
	var req Request
	c.BodyParser(&req)

	sess, _ := h.store.Get(c)
	reporterID := sess.Get("user_id").(string)

	checkBody, err := h.db.Select("user_reports",
		fmt.Sprintf("reporter_id=eq.%s&reported_user_id=eq.%s&select=id", reporterID, req.UserID), true)
	if err == nil {
		var existing []interface{}
		json.Unmarshal(checkBody, &existing)
		if len(existing) > 0 {
			return c.JSON(fiber.Map{"error": "Vous avez déjà signalé cet utilisateur"})
		}
	}

	descJSON := "null"
	if req.Description != "" {
		descJSON = `"` + strings.ReplaceAll(req.Description, `"`, `\"`) + `"`
	}

	insertData := fmt.Sprintf(`{"reporter_id":"%s","reported_user_id":"%s","reason":"%s","description":%s}`, reporterID, req.UserID, req.Reason, descJSON)
	_, err = h.db.Insert("user_reports", []byte(insertData), true)
	if err != nil {
		return c.JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true})
}

func (h *Handler) APIGetFriendshipStatus(c *fiber.Ctx) error {
	targetID := c.Query("user_id")
	sess, _ := h.store.Get(c)
	userID := sess.Get("user_id").(string)

	body, err := h.db.Select("friendships",
		fmt.Sprintf("or=(and(requester_id.eq.%s,addressee_id.eq.%s),and(requester_id.eq.%s,addressee_id.eq.%s))&select=id,status,requester_id",
			userID, targetID, targetID, userID), true)
	if err != nil {
		return c.JSON(fiber.Map{"status": nil, "friendship_id": nil, "is_requester": false})
	}
	var friendships []map[string]interface{}
	json.Unmarshal(body, &friendships)
	if len(friendships) > 0 {
		f := friendships[0]
		status, _ := f["status"].(string)
		friendshipID, _ := f["id"].(string)
		requesterID, _ := f["requester_id"].(string)
		isRequester := requesterID == userID
		return c.JSON(fiber.Map{
			"status":        status,
			"friendship_id": friendshipID,
			"is_requester":  isRequester,
		})
	}

	return c.JSON(fiber.Map{"status": nil, "friendship_id": nil, "is_requester": false})
}

func (h *Handler) APIDeleteChallenge(c *fiber.Ctx) error {
	type Request struct {
		SessionID string `json:"session_id"`
	}
	var req Request
	c.BodyParser(&req)

	sess, _ := h.store.Get(c)
	userID := sess.Get("user_id").(string)

	sBody, err := h.db.Select("challenge_sessions",
		fmt.Sprintf("id=eq.%s&select=creator_id,status,total_xp_pool", req.SessionID), true)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Défi introuvable"})
	}
	var rows []map[string]interface{}
	json.Unmarshal(sBody, &rows)
	if len(rows) == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Défi introuvable"})
	}

	session := rows[0]
	creatorID, _ := session["creator_id"].(string)
	if creatorID != userID {
		return c.Status(403).JSON(fiber.Map{"error": "Seul le créateur peut supprimer"})
	}

	status, _ := session["status"].(string)
	if status == "completed" || status == "cancelled" {
		h.db.Update("challenge_sessions", fmt.Sprintf("id=eq.%s", req.SessionID),
			[]byte(`{"status":"cancelled"}`), true)
		h.db.Delete("challenge_invitations", fmt.Sprintf("session_id=eq.%s", req.SessionID), true)
		return c.JSON(fiber.Map{"success": true})
	}
	if status != "waiting" {
		return c.Status(400).JSON(fiber.Map{"error": "Impossible de supprimer un défi en cours"})
	}

	partBody, _ := h.db.Select("challenge_participants",
		fmt.Sprintf("session_id=eq.%s&select=user_id,xp_bet,individual_bet", req.SessionID), true)
	var participants []map[string]interface{}
	json.Unmarshal(partBody, &participants)
	for _, p := range participants {
		pUID, _ := p["user_id"].(string)
		if pUID == creatorID {
			continue
		}
		xpBet := database.DBInt(p["xp_bet"])
		indBet := database.DBInt(p["individual_bet"])
		if indBet > xpBet {
			return c.Status(400).JSON(fiber.Map{"error": "Impossible de supprimer : un participant a renchéri"})
		}
	}

	h.db.Update("challenge_sessions", fmt.Sprintf("id=eq.%s", req.SessionID),
		[]byte(`{"status":"cancelled"}`), true)

	h.db.Delete("challenge_invitations", fmt.Sprintf("session_id=eq.%s", req.SessionID), true)

	for _, p := range participants {
		pUID, _ := p["user_id"].(string)
		xpBet := database.DBInt(p["xp_bet"])
		if pUID == creatorID {
			continue
		}
		profileBody, _ := h.db.Select("user_profiles", fmt.Sprintf("id=eq.%s&select=xp", pUID), true)
		var pRows []map[string]interface{}
		json.Unmarshal(profileBody, &pRows)
		if len(pRows) > 0 {
			currentXP := int(pRows[0]["xp"].(float64))
			newXP := currentXP + xpBet
			newRank := getRankForXP(newXP, h.loadRanks())
			xpData, _ := json.Marshal(map[string]interface{}{"xp": newXP, "rank": newRank})
			h.db.Update("user_profiles", fmt.Sprintf("id=eq.%s", pUID), xpData, true)
			createNotification(h.db, pUID, "challenge_invitation", "🔄 Défi annulé",
				fmt.Sprintf("Le défi a été annulé. %d XP remboursés.", xpBet))
		}
	}

	return c.JSON(fiber.Map{"success": true})
}

func (h *Handler) APIInviteToChallenge(c *fiber.Ctx) error {
	type Request struct {
		SessionID string   `json:"session_id"`
		FriendIDs []string `json:"friend_ids"`
	}
	var req Request
	c.BodyParser(&req)

	sess, _ := h.store.Get(c)
	userID := sess.Get("user_id").(string)

	sBody, err := h.db.Select("challenge_sessions",
		fmt.Sprintf("id=eq.%s&select=creator_id,status", req.SessionID), true)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Défi introuvable"})
	}
	var rows []map[string]interface{}
	json.Unmarshal(sBody, &rows)
	if len(rows) == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Défi introuvable"})
	}
	creatorID, _ := rows[0]["creator_id"].(string)
	if creatorID != userID {
		return c.Status(403).JSON(fiber.Map{"error": "Seul le créateur peut inviter"})
	}
	status, _ := rows[0]["status"].(string)
	if status != "waiting" {
		return c.Status(400).JSON(fiber.Map{"error": "On ne peut inviter que sur un défi en attente"})
	}

	partCheckBody, _ := h.db.Select("challenge_participants",
		fmt.Sprintf("session_id=eq.%s&status=eq.accepted&select=bet_accepted,user_id", req.SessionID), true)
	var partChecks []map[string]interface{}
	json.Unmarshal(partCheckBody, &partChecks)
	hasNonCreator := false
	allAccepted := len(partChecks) > 0
	for _, p := range partChecks {
		pUID, _ := p["user_id"].(string)
		if pUID != creatorID {
			hasNonCreator = true
		}
		if !database.DBBool(p["bet_accepted"]) {
			allAccepted = false
			break
		}
	}
	if allAccepted && hasNonCreator {
		return c.Status(400).JSON(fiber.Map{"error": "Tous les participants ont déjà accepté la mise, le défi est sur le point de commencer"})
	}

	inviterName := "Quelqu'un"
	pBody, _ := h.db.Select("user_profiles",
		fmt.Sprintf("id=eq.%s&select=username,nickname", userID), false)
	var pRows []map[string]interface{}
	json.Unmarshal(pBody, &pRows)
	if len(pRows) > 0 {
		if nn, ok := pRows[0]["nickname"].(string); ok && nn != "" {
			inviterName = nn
		} else if un, ok := pRows[0]["username"].(string); ok {
			inviterName = un
		}
	}

	quizTitle := "Quiz"
	qBody, _ := h.db.Select("challenge_sessions",
		fmt.Sprintf("id=eq.%s&select=quiz:quiz_id(title)", req.SessionID), true)
	var qRows []map[string]interface{}
	json.Unmarshal(qBody, &qRows)
	if len(qRows) > 0 {
		if quiz, ok := qRows[0]["quiz"].(map[string]interface{}); ok {
			if t, ok := quiz["title"].(string); ok {
				quizTitle = t
			}
		}
	}

	xpBet := 0
	xpBody, _ := h.db.Select("challenge_sessions",
		fmt.Sprintf("id=eq.%s&select=total_xp_pool", req.SessionID), true)
	var xpRows []map[string]interface{}
	json.Unmarshal(xpBody, &xpRows)
	if len(xpRows) > 0 {
		xpBet = int(xpRows[0]["total_xp_pool"].(float64))
	}

	invited := 0
	for _, friendID := range req.FriendIDs {
		friendID = strings.TrimSpace(friendID)
		if friendID == "" || friendID == userID {
			continue
		}

		checkBody, _ := h.db.Select("challenge_invitations",
			fmt.Sprintf("session_id=eq.%s&invitee_id=eq.%s&select=id", req.SessionID, friendID), true)
		var existing []map[string]interface{}
		json.Unmarshal(checkBody, &existing)
		if len(existing) > 0 {
			status, _ := existing[0]["status"].(string)
			if status == "pending" || status == "accepted" {
				continue
			}
			h.db.Update("challenge_invitations",
				fmt.Sprintf("session_id=eq.%s&invitee_id=eq.%s", req.SessionID, friendID),
				[]byte(`{"status":"pending"}`), true)
		} else {
			invData, _ := json.Marshal(map[string]interface{}{
				"session_id": req.SessionID,
				"inviter_id": userID,
				"invitee_id": friendID,
				"status":     "pending",
				"expires_at": time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339),
			})
			h.db.Insert("challenge_invitations", invData, true)
		}

		createNotification(h.db, friendID, "challenge_invitation",
			"⚔️ Défi reçu !",
			fmt.Sprintf("%s vous a défié sur « %s » ! Mise : %d XP", inviterName, quizTitle, xpBet),
			fmt.Sprintf(`{"session_id":"%s","inviter_id":"%s"}`, req.SessionID, userID))
		h.sendSystemMessage(userID, friendID,
			fmt.Sprintf("⚔️ %s vous a défié sur « %s » ! Mise : %d XP. Acceptez dans l'onglet Défis.", inviterName, quizTitle, xpBet))
		invited++
	}

	return c.JSON(fiber.Map{"success": true, "invited": invited})
}

func (h *Handler) APIResendChallengeInvitation(c *fiber.Ctx) error {
	type Request struct {
		InvitationID string `json:"invitation_id"`
	}
	var req Request
	c.BodyParser(&req)

	sess, _ := h.store.Get(c)
	userID := sess.Get("user_id").(string)

	invBody, err := h.db.Select("challenge_invitations",
		fmt.Sprintf("id=eq.%s&select=session_id,inviter_id,invitee_id", req.InvitationID), true)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Introuvable"})
	}
	var invRows []map[string]interface{}
	json.Unmarshal(invBody, &invRows)
	if len(invRows) == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Introuvable"})
	}

	inv := invRows[0]
	inviterID, _ := inv["inviter_id"].(string)
	if inviterID != userID {
		return c.Status(403).JSON(fiber.Map{"error": "Non autorisé"})
	}

	inviteeID, _ := inv["invitee_id"].(string)
	sessionID, _ := inv["session_id"].(string)

	h.db.Update("challenge_invitations",
		fmt.Sprintf("id=eq.%s", req.InvitationID),
		[]byte(`{"status":"pending"}`), true)

	inviterName := "Quelqu'un"
	pBody, _ := h.db.Select("user_profiles",
		fmt.Sprintf("id=eq.%s&select=username,nickname", userID), false)
	var pRows []map[string]interface{}
	json.Unmarshal(pBody, &pRows)
	if len(pRows) > 0 {
		if nn, ok := pRows[0]["nickname"].(string); ok && nn != "" {
			inviterName = nn
		} else if un, ok := pRows[0]["username"].(string); ok {
			inviterName = un
		}
	}

	quizTitle := "Quiz"
	qBody, _ := h.db.Select("challenge_sessions",
		fmt.Sprintf("id=eq.%s&select=quiz:quiz_id(title),total_xp_pool", sessionID), true)
	var qRows []map[string]interface{}
	json.Unmarshal(qBody, &qRows)
	xpBet := 0
	if len(qRows) > 0 {
		if quiz, ok := qRows[0]["quiz"].(map[string]interface{}); ok {
			if t, ok := quiz["title"].(string); ok {
				quizTitle = t
			}
		}
		xpBet = int(qRows[0]["total_xp_pool"].(float64))
	}

	createNotification(h.db, inviteeID, "challenge_invitation",
		"⚔️ Nouvelle invitation !",
		fmt.Sprintf("%s vous a relancé pour un défi sur « %s » ! Mise : %d XP", inviterName, quizTitle, xpBet),
		fmt.Sprintf(`{"session_id":"%s","inviter_id":"%s"}`, sessionID, userID))
	h.sendSystemMessage(userID, inviteeID,
		fmt.Sprintf("⚔️ %s vous relance pour le défi « %s » ! Mise : %d XP.", inviterName, quizTitle, xpBet))

	return c.JSON(fiber.Map{"success": true})
}

func (h *Handler) APIRaiseBet(c *fiber.Ctx) error {
	type Request struct {
		SessionID      string `json:"session_id"`
		IncreaseAmount int    `json:"increase_amount"`
	}
	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Données invalides"})
	}
	if req.SessionID == "" || req.IncreaseAmount <= 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Montant invalide"})
	}

	user := c.Locals("user").(*UserProfile)
	userID := user.ID

	pBody, err := h.db.Select("challenge_participants",
		fmt.Sprintf("session_id=eq.%s&user_id=eq.%s&status=eq.accepted&select=individual_bet", req.SessionID, userID), true)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Erreur"})
	}
	var parts []map[string]interface{}
	json.Unmarshal(pBody, &parts)

	if len(parts) == 0 {
		invBody, _ := h.db.Select("challenge_invitations",
			fmt.Sprintf("session_id=eq.%s&invitee_id=eq.%s&status=eq.pending", req.SessionID, userID), true)
		var invs []map[string]interface{}
		json.Unmarshal(invBody, &invs)
		if len(invs) == 0 {
			return c.Status(403).JSON(fiber.Map{"error": "Vous n'êtes pas participant"})
		}
		h.db.Update("challenge_invitations", fmt.Sprintf("id=eq.%s", invs[0]["id"]),
			[]byte(`{"status":"accepted"}`), true)
		sPool, _ := h.db.Select("challenge_sessions",
			fmt.Sprintf("id=eq.%s&select=total_xp_pool", req.SessionID), true)
		var spRows []map[string]interface{}
		json.Unmarshal(sPool, &spRows)
		totalPool := 100
		if len(spRows) > 0 {
			if p, ok := spRows[0]["total_xp_pool"].(float64); ok {
				totalPool = int(p)
			}
		}
		inviteeBet := totalPool
		h.db.Update("challenge_sessions", fmt.Sprintf("id=eq.%s", req.SessionID),
			[]byte(fmt.Sprintf(`{"total_xp_pool":%d}`, totalPool+inviteeBet)), true)
		partInsert := fmt.Sprintf(`{"session_id":"%s","user_id":"%s","xp_bet":%d,"status":"accepted","individual_bet":%d,"bet_version":1,"bet_accepted":false}`,
			req.SessionID, userID, inviteeBet, inviteeBet)
		h.db.Insert("challenge_participants", []byte(partInsert), true)
		pBody, _ = h.db.Select("challenge_participants",
			fmt.Sprintf("session_id=eq.%s&user_id=eq.%s&status=eq.accepted&select=individual_bet", req.SessionID, userID), true)
		json.Unmarshal(pBody, &parts)
		if len(parts) == 0 {
			return c.Status(500).JSON(fiber.Map{"error": "Erreur lors de l'acceptation"})
		}
	}

	currentBet := database.DBInt(parts[0]["individual_bet"])
	newTotalBet := currentBet + req.IncreaseAmount
	if user.XP < newTotalBet {
		return c.Status(400).JSON(fiber.Map{
			"error":           fmt.Sprintf("XP insuffisant. Vous avez %d XP, votre mise totale serait de %d XP", user.XP, newTotalBet),
			"current_xp":      user.XP,
			"required_xp":     newTotalBet,
			"increase_amount": req.IncreaseAmount,
		})
	}

	allBody, _ := h.db.Select("challenge_participants",
		fmt.Sprintf("session_id=eq.%s&status=eq.accepted&select=user_id,individual_bet", req.SessionID), true)
	var allParts []map[string]interface{}
	json.Unmarshal(allBody, &allParts)

	participantCount := len(allParts)
	for _, p := range allParts {
		pid := database.DBValue(p["user_id"])
		pBet := database.DBInt(p["individual_bet"])
		newPBet := pBet + req.IncreaseAmount
		if pid == userID {
			h.db.Update("challenge_participants",
				fmt.Sprintf("session_id=eq.%s&user_id=eq.%s", req.SessionID, pid),
				[]byte(fmt.Sprintf(`{"individual_bet":%d,"bet_accepted":true}`, newPBet)), true)
		} else {
			h.db.Update("challenge_participants",
				fmt.Sprintf("session_id=eq.%s&user_id=eq.%s", req.SessionID, pid),
				[]byte(fmt.Sprintf(`{"individual_bet":%d,"bet_accepted":false}`, newPBet)), true)
		}
	}

	sBody, _ := h.db.Select("challenge_sessions",
		fmt.Sprintf("id=eq.%s&select=total_xp_pool,quiz:quiz_id(title)", req.SessionID), true)
	var sessions []map[string]interface{}
	json.Unmarshal(sBody, &sessions)
	currentPool := 0
	quizTitle := "Quiz"
	if len(sessions) > 0 {
		currentPool = database.DBInt(sessions[0]["total_xp_pool"])
		if quiz, ok := sessions[0]["quiz"].(map[string]interface{}); ok {
			if t, ok := quiz["title"].(string); ok {
				quizTitle = t
			}
		}
	}
	newPool := currentPool + (req.IncreaseAmount * participantCount)
	h.db.Update("challenge_sessions",
		fmt.Sprintf("id=eq.%s", req.SessionID),
		[]byte(fmt.Sprintf(`{"total_xp_pool":%d}`, newPool)), true)

	raiserName := user.Username
	if user.Nickname != nil && *user.Nickname != "" {
		raiserName = *user.Nickname
	}
	for _, p := range allParts {
		oid := database.DBValue(p["user_id"])
		if oid == userID {
			continue
		}
		createNotification(h.db, oid, "challenge_bet_raise",
			"⚔️ Mise renchérie !",
			fmt.Sprintf("%s a augmenté la mise du défi. Mise de chacun : %d XP. Acceptez-vous ?", raiserName, database.DBInt(p["individual_bet"])+req.IncreaseAmount),
			fmt.Sprintf(`{"session_id":"%s","raiser_id":"%s","new_pool":%d}`, req.SessionID, userID, newPool))
		h.sendSystemMessage(userID, oid,
			fmt.Sprintf("⚔️ %s a augmenté la mise du défi « %s ». Chaque participant mise maintenant %d XP. Répondez dans la page du défi.", raiserName, quizTitle, database.DBInt(p["individual_bet"])+req.IncreaseAmount))
	}

	return c.JSON(fiber.Map{
		"success":      true,
		"new_pool":     newPool,
		"total_xp_pool": newPool,
	})
}

func (h *Handler) APIBetRespond(c *fiber.Ctx) error {
	type Request struct {
		SessionID string `json:"session_id"`
		Accept    bool   `json:"accept"`
	}
	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Données invalides"})
	}
	if req.SessionID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "ID requis"})
	}

	user := c.Locals("user").(*UserProfile)
	userID := user.ID

	pBody, err := h.db.Select("challenge_participants",
		fmt.Sprintf("session_id=eq.%s&user_id=eq.%s&status=eq.accepted&select=individual_bet,bet_accepted", req.SessionID, userID), true)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Erreur"})
	}
	var parts []map[string]interface{}
	json.Unmarshal(pBody, &parts)

	if len(parts) == 0 {
		invBody, _ := h.db.Select("challenge_invitations",
			fmt.Sprintf("session_id=eq.%s&invitee_id=eq.%s&status=eq.pending", req.SessionID, userID), true)
		var invs []map[string]interface{}
		json.Unmarshal(invBody, &invs)
		if len(invs) == 0 {
			return c.Status(403).JSON(fiber.Map{"error": "Vous n'êtes pas participant"})
		}
		if req.Accept {
			h.db.Update("challenge_invitations", fmt.Sprintf("id=eq.%s", invs[0]["id"]),
				[]byte(`{"status":"accepted"}`), true)
			sPool, _ := h.db.Select("challenge_sessions",
				fmt.Sprintf("id=eq.%s&select=total_xp_pool", req.SessionID), true)
			var spRows []map[string]interface{}
			json.Unmarshal(sPool, &spRows)
			totalPool := 100
			if len(spRows) > 0 {
				if p, ok := spRows[0]["total_xp_pool"].(float64); ok {
					totalPool = int(p)
				}
			}
			inviteeBet := totalPool
			h.db.Update("challenge_sessions", fmt.Sprintf("id=eq.%s", req.SessionID),
				[]byte(fmt.Sprintf(`{"total_xp_pool":%d}`, totalPool+inviteeBet)), true)
			partInsert := fmt.Sprintf(`{"session_id":"%s","user_id":"%s","xp_bet":%d,"status":"accepted","individual_bet":%d,"bet_version":1,"bet_accepted":false}`,
				req.SessionID, userID, inviteeBet, inviteeBet)
			h.db.Insert("challenge_participants", []byte(partInsert), true)
			pBody, _ = h.db.Select("challenge_participants",
				fmt.Sprintf("session_id=eq.%s&user_id=eq.%s&status=eq.accepted&select=individual_bet,bet_accepted", req.SessionID, userID), true)
			json.Unmarshal(pBody, &parts)
			if len(parts) == 0 {
				return c.Status(500).JSON(fiber.Map{"error": "Erreur lors de l'acceptation"})
			}
			p := parts[0]
			individualBet := database.DBInt(p["individual_bet"])
			if user.XP < individualBet {
				return c.Status(400).JSON(fiber.Map{
					"error":       fmt.Sprintf("XP insuffisant pour accepter cette mise. Vous avez %d XP, votre mise est de %d XP", user.XP, individualBet),
					"current_xp":  user.XP,
					"required_xp": individualBet,
				})
			}
			h.db.Update("challenge_participants",
				fmt.Sprintf("session_id=eq.%s&user_id=eq.%s", req.SessionID, userID),
				[]byte(`{"bet_accepted":true}`), true)
			h.checkChallengeConsensus(req.SessionID)
			return c.JSON(fiber.Map{"success": true})
		} else {
			h.db.Update("challenge_invitations", fmt.Sprintf("id=eq.%s", invs[0]["id"]),
				[]byte(`{"status":"refused"}`), true)
			return c.JSON(fiber.Map{"success": true})
		}
	}

	p := parts[0]
	individualBet := database.DBInt(p["individual_bet"])
	betAccepted := database.DBBool(p["bet_accepted"])

	if betAccepted {
		return c.JSON(fiber.Map{"success": true, "message": "Déjà accepté"})
	}

	if req.Accept {
		if user.XP < individualBet {
			return c.Status(400).JSON(fiber.Map{
				"error":       fmt.Sprintf("XP insuffisant pour accepter cette mise. Vous avez %d XP, votre mise est de %d XP", user.XP, individualBet),
				"current_xp":  user.XP,
				"required_xp": individualBet,
			})
		}
		h.db.Update("challenge_participants",
			fmt.Sprintf("session_id=eq.%s&user_id=eq.%s", req.SessionID, userID),
			[]byte(`{"bet_accepted":true}`), true)

		sBody, _ := h.db.Select("challenge_sessions",
			fmt.Sprintf("id=eq.%s&select=quiz:quiz_id(title)", req.SessionID), true)
		quizTitle := "Quiz"
		var sess []map[string]interface{}
		json.Unmarshal(sBody, &sess)
		if len(sess) > 0 {
			if quiz, ok := sess[0]["quiz"].(map[string]interface{}); ok {
				if t, ok := quiz["title"].(string); ok {
					quizTitle = t
				}
			}
		}

		acceptorName := user.Username
		if user.Nickname != nil && *user.Nickname != "" {
			acceptorName = *user.Nickname
		}

		otherBody, _ := h.db.Select("challenge_participants",
			fmt.Sprintf("session_id=eq.%s&status=eq.accepted&select=user_id", req.SessionID), true)
		var others []map[string]interface{}
		json.Unmarshal(otherBody, &others)
		for _, o := range others {
			oid := database.DBValue(o["user_id"])
			if oid == userID {
				continue
			}
			createNotification(h.db, oid, "challenge_bet_accepted",
				"✅ Mise acceptée",
				fmt.Sprintf("%s a accepté la mise du défi « %s ».", acceptorName, quizTitle),
				fmt.Sprintf(`{"session_id":"%s"}`, req.SessionID))
		}

		h.checkChallengeConsensus(req.SessionID)
	} else {
		refuserName := user.Username
		if user.Nickname != nil && *user.Nickname != "" {
			refuserName = *user.Nickname
		}
		h.db.Update("challenge_participants",
			fmt.Sprintf("session_id=eq.%s&user_id=eq.%s", req.SessionID, userID),
			[]byte(`{"status":"left"}`), true)

		poolBody, _ := h.db.Select("challenge_sessions",
			fmt.Sprintf("id=eq.%s&select=total_xp_pool", req.SessionID), true)
		var poolRows []map[string]interface{}
		json.Unmarshal(poolBody, &poolRows)
		if len(poolRows) > 0 {
			pool := int(poolRows[0]["total_xp_pool"].(float64))
			newPool := pool - individualBet
			if newPool < 0 {
				newPool = 0
			}
			h.db.Update("challenge_sessions",
				fmt.Sprintf("id=eq.%s", req.SessionID),
				[]byte(fmt.Sprintf(`{"total_xp_pool":%d}`, newPool)), true)
		}

		sBody, _ := h.db.Select("challenge_sessions",
			fmt.Sprintf("id=eq.%s&select=quiz:quiz_id(title)", req.SessionID), true)
		quizTitle := "Quiz"
		var sess []map[string]interface{}
		json.Unmarshal(sBody, &sess)
		if len(sess) > 0 {
			if quiz, ok := sess[0]["quiz"].(map[string]interface{}); ok {
				if t, ok := quiz["title"].(string); ok {
					quizTitle = t
				}
			}
		}

		otherBody, _ := h.db.Select("challenge_participants",
			fmt.Sprintf("session_id=eq.%s&status=eq.accepted&select=user_id", req.SessionID), true)
		var others []map[string]interface{}
		json.Unmarshal(otherBody, &others)
		for _, o := range others {
			oid := database.DBValue(o["user_id"])
			if oid == userID {
				continue
			}
			createNotification(h.db, oid, "challenge_bet_refused",
				"❌ Mise refusée",
				fmt.Sprintf("%s a refusé la mise et quitté le défi « %s ».", refuserName, quizTitle),
				fmt.Sprintf(`{"session_id":"%s"}`, req.SessionID))
			h.sendSystemMessage(userID, oid,
				fmt.Sprintf("❌ %s a refusé la mise et quitté le défi « %s ».", refuserName, quizTitle))
		}

		if len(others) <= 1 {
			h.db.Update("challenge_sessions",
				fmt.Sprintf("id=eq.%s", req.SessionID),
				[]byte(`{"status":"cancelled"}`), true)
			for _, o := range others {
				oid := database.DBValue(o["user_id"])
				createNotification(h.db, oid, "challenge_cancelled",
					"🗑️ Défi annulé",
					"Le défi a été annulé car il ne reste qu'un seul participant.",
					fmt.Sprintf(`{"session_id":"%s"}`, req.SessionID))
			}
			createNotification(h.db, userID, "challenge_cancelled",
				"🗑️ Défi annulé",
				"Le défi a été annulé car il ne reste qu'un seul participant.",
				fmt.Sprintf(`{"session_id":"%s"}`, req.SessionID))
		}
	}

	return c.JSON(fiber.Map{"success": true})
}

func (h *Handler) checkChallengeConsensus(sessionID string) {
	acceptedBody, _ := h.db.Select("challenge_participants",
		fmt.Sprintf("session_id=eq.%s&status=eq.accepted&select=individual_bet,bet_accepted,user_id", sessionID), true)
	var accepted []map[string]interface{}
	json.Unmarshal(acceptedBody, &accepted)

	if len(accepted) < 2 {
		return
	}

	allAgreed := true
	var xpMap []map[string]interface{}
	for _, p := range accepted {
		if !database.DBBool(p["bet_accepted"]) {
			allAgreed = false
			break
		}
		xpMap = append(xpMap, p)
	}

	if !allAgreed {
		return
	}

	profiles := []map[string]interface{}{}
	for _, p := range xpMap {
		uid := database.DBValue(p["user_id"])
		indBet := database.DBInt(p["individual_bet"])
		pBody, _ := h.db.Select("user_profiles",
			fmt.Sprintf("id=eq.%s&select=xp", uid), true)
		var uRows []map[string]interface{}
		json.Unmarshal(pBody, &uRows)
		if len(uRows) > 0 {
			userXP := int(uRows[0]["xp"].(float64))
			profiles = append(profiles, map[string]interface{}{
				"user_id":  uid,
				"xp":       userXP,
				"ind_bet":  indBet,
				"sufficient": userXP >= indBet,
			})
		}
	}

	insufficientCount := 0
	totalDeduct := 0
	for _, pr := range profiles {
		if !pr["sufficient"].(bool) {
			insufficientCount++
			uid := pr["user_id"].(string)
			indBet := pr["ind_bet"].(int)
			totalDeduct += indBet
			h.db.Update("challenge_participants",
				fmt.Sprintf("session_id=eq.%s&user_id=eq.%s", sessionID, uid),
				[]byte(`{"status":"left"}`), true)
			createNotification(h.db, uid, "challenge_bet_insufficient",
				"❌ XP insuffisant",
				"Vous n'avez plus assez d'XP pour votre mise. Vous avez été retiré du défi.",
				fmt.Sprintf(`{"session_id":"%s"}`, sessionID))
		}
	}

	if insufficientCount > 0 {
		poolBody, _ := h.db.Select("challenge_sessions",
			fmt.Sprintf("id=eq.%s&select=total_xp_pool", sessionID), true)
		var poolRows []map[string]interface{}
		json.Unmarshal(poolBody, &poolRows)
		if len(poolRows) > 0 {
			pool := int(poolRows[0]["total_xp_pool"].(float64))
			newPool := pool - totalDeduct
			if newPool < 0 {
				newPool = 0
			}
			h.db.Update("challenge_sessions",
				fmt.Sprintf("id=eq.%s", sessionID),
				[]byte(fmt.Sprintf(`{"total_xp_pool":%d}`, newPool)), true)
		}

		remainingBody, _ := h.db.Select("challenge_participants",
			fmt.Sprintf("session_id=eq.%s&status=eq.accepted&select=user_id", sessionID), true)
		var remaining []map[string]interface{}
		json.Unmarshal(remainingBody, &remaining)
		if len(remaining) <= 1 {
			h.db.Update("challenge_sessions",
				fmt.Sprintf("id=eq.%s", sessionID),
				[]byte(`{"status":"cancelled"}`), true)
			for _, r := range remaining {
				rid := database.DBValue(r["user_id"])
				createNotification(h.db, rid, "challenge_cancelled",
					"🗑️ Défi annulé",
					"Le défi a été annulé car il ne reste qu'un seul participant.",
					fmt.Sprintf(`{"session_id":"%s"}`, sessionID))
			}
		}
		return
	}

	h.deductXPForChallenge(sessionID)

	sBody, _ := h.db.Select("challenge_sessions",
		fmt.Sprintf("id=eq.%s&select=quiz:quiz_id(title)", sessionID), true)
	quizTitle := "Quiz"
	var sess []map[string]interface{}
	json.Unmarshal(sBody, &sess)
	if len(sess) > 0 {
		if quiz, ok := sess[0]["quiz"].(map[string]interface{}); ok {
			if t, ok := quiz["title"].(string); ok {
				quizTitle = t
			}
		}
	}

	for _, pr := range profiles {
		uid := pr["user_id"].(string)
		createNotification(h.db, uid, "challenge_ready",
			"⚔️ Défi prêt !",
			fmt.Sprintf("Tout le monde est d'accord ! Le défi « %s » peut commencer. %d XP ont été débités.", quizTitle, pr["ind_bet"].(int)),
			fmt.Sprintf(`{"session_id":"%s"}`, sessionID))
	}
}

func (h *Handler) deductXPForChallenge(sessionID string) {
	acceptedBody, _ := h.db.Select("challenge_participants",
		fmt.Sprintf("session_id=eq.%s&status=eq.accepted&select=user_id,individual_bet", sessionID), true)
	var accepted []map[string]interface{}
	json.Unmarshal(acceptedBody, &accepted)

	for _, p := range accepted {
		uid := database.DBValue(p["user_id"])
		indBet := database.DBInt(p["individual_bet"])
		if indBet <= 0 {
			continue
		}
		txData := fmt.Sprintf(`{"user_id":"%s","amount":%d,"type":"challenge_bet","description":"Mise défi","reference_id":"%s","created_at":"%s"}`,
			uid, -indBet, sessionID, time.Now().UTC().Format(time.RFC3339))
		h.db.Insert("xp_transactions", []byte(txData), true)

		pBody, _ := h.db.Select("user_profiles", fmt.Sprintf("id=eq.%s&select=xp", uid), true)
		var pRows []map[string]interface{}
		json.Unmarshal(pBody, &pRows)
		if len(pRows) > 0 {
			currentXP := int(pRows[0]["xp"].(float64))
			newXP := currentXP - indBet
			if newXP < 0 {
				newXP = 0
			}
			newRank := getRankForXP(newXP, h.loadRanks())
			xpData, _ := json.Marshal(map[string]interface{}{"xp": newXP, "rank": newRank})
			h.db.Update("user_profiles", fmt.Sprintf("id=eq.%s", uid), xpData, true)
		}
	}

	h.db.Update("challenge_sessions",
		fmt.Sprintf("id=eq.%s", sessionID),
		[]byte(`{"status":"ready"}`), true)
}

func (h *Handler) APIGetChallengeDetail(c *fiber.Ctx) error {
	sessionID := c.Params("id")
	user := c.Locals("user").(*UserProfile)
	userID := user.ID

	sBody, err := h.db.Select("challenge_sessions",
		fmt.Sprintf("id=eq.%s&select=*,quiz:quiz_id(title,question_count),creator:creator_id(username,nickname)", sessionID), true)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Défi introuvable"})
	}
	var sessions []map[string]interface{}
	json.Unmarshal(sBody, &sessions)
	if len(sessions) == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Défi introuvable"})
	}
	session := sessions[0]

	partBody, _ := h.db.Select("challenge_participants",
		fmt.Sprintf("session_id=eq.%s&select=*,user:user_id(username,nickname,avatar_url,rank)", sessionID), true)
	var participants []map[string]interface{}
	json.Unmarshal(partBody, &participants)

	invBody, _ := h.db.Select("challenge_invitations",
		fmt.Sprintf("session_id=eq.%s&select=*,invitee:invitee_id(username,nickname)", sessionID), true)
	var invitations []map[string]interface{}
	json.Unmarshal(invBody, &invitations)

	scoreBody, _ := h.db.Select("challenge_scores",
		fmt.Sprintf("session_id=eq.%s&user_id=eq.%s&select=correct_count,total_questions,time_taken_ms", sessionID, userID), true)
	var scoreRows []map[string]interface{}
	json.Unmarshal(scoreBody, &scoreRows)
	var userScore interface{}
	if len(scoreRows) > 0 {
		userScore = scoreRows[0]
	}

	return c.JSON(fiber.Map{
		"session":      session,
		"participants": participants,
		"invitations":  invitations,
		"user_score":   userScore,
	})
}

func (h *Handler) APIGetAllSuggestions(c *fiber.Ctx) error {
	userID := c.Locals("user").(*UserProfile).ID

	sugBody, err := h.db.Select("forum_suggestions",
		"select=id,channel_id,title,content,week_label,user_id,created_at&order=created_at.desc", true)
	if err != nil {
		log.Printf("[APIGetAllSuggestions] Select error: %v", err)
		return c.JSON(fiber.Map{"suggestions": []interface{}{}})
	}
	log.Printf("[APIGetAllSuggestions] raw body (%d bytes): %s", len(sugBody), string(sugBody))
	var suggestions []map[string]interface{}
	json.Unmarshal(sugBody, &suggestions)
	log.Printf("[APIGetAllSuggestions] user=%s found %d suggestions", userID, len(suggestions))
	for i, s := range suggestions {
		log.Printf("[APIGetAllSuggestions] suggestion[%d]: id=%v title=%v", i, s["id"], s["title"])
	}

	userIDs := map[string]bool{}
	suggestionIDs := []string{}
	for _, s := range suggestions {
		if uid, ok := s["user_id"].(string); ok {
			userIDs[uid] = true
		}
		if sid, ok := s["id"].(string); ok {
			suggestionIDs = append(suggestionIDs, sid)
		}
	}
	uidList := make([]string, 0, len(userIDs))
	for uid := range userIDs {
		uidList = append(uidList, uid)
	}
	profiles := map[string]map[string]interface{}{}
	if len(uidList) > 0 {
		pList, _ := h.db.GetProfiles(uidList)
		for _, p := range pList {
			if pid, ok := p["id"].(string); ok {
				profiles[pid] = p
			}
		}
	}

	voteMap := map[string]string{}
	upCounts := map[string]int{}
	downCounts := map[string]int{}
	if len(suggestionIDs) > 0 {
		votesBody, _ := h.db.Select("forum_suggestion_votes",
			fmt.Sprintf("user_id=eq.%s&select=suggestion_id,vote_type", userID), false)
		var myVotes []map[string]interface{}
		json.Unmarshal(votesBody, &myVotes)
		for _, v := range myVotes {
			sid, _ := v["suggestion_id"].(string)
			vt, _ := v["vote_type"].(string)
			voteMap[sid] = vt
		}

		for _, sid := range suggestionIDs {
			upBody, _ := h.db.Select("forum_suggestion_votes",
				fmt.Sprintf("suggestion_id=eq.%s&vote_type=eq.up&select=id", sid), true)
			var ups []map[string]interface{}
			json.Unmarshal(upBody, &ups)
			upCounts[sid] = len(ups)

			downBody, _ := h.db.Select("forum_suggestion_votes",
				fmt.Sprintf("suggestion_id=eq.%s&vote_type=eq.down&select=id", sid), true)
			var downs []map[string]interface{}
			json.Unmarshal(downBody, &downs)
			downCounts[sid] = len(downs)
		}
	}

	result := make([]map[string]interface{}, 0, len(suggestions))
	for _, s := range suggestions {
		sid, _ := s["id"].(string)
		uid, _ := s["user_id"].(string)
		item := map[string]interface{}{
			"id":          s["id"],
			"title":       s["title"],
			"content":     s["content"],
			"up_votes":    upCounts[sid],
			"down_votes":  downCounts[sid],
			"week_label":  s["week_label"],
			"user_id":     s["user_id"],
			"created_at":  s["created_at"],
			"my_vote":     voteMap[sid],
			"user":        profiles[uid],
		}
		result = append(result, item)
	}

	return c.JSON(fiber.Map{"suggestions": result})
}

func (h *Handler) APISetRewardMode(c *fiber.Ctx) error {
	user := c.Locals("user").(*UserProfile)
	var req struct {
		SessionID string `json:"session_id"`
		Mode      string `json:"mode"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Données invalides"})
	}
	if req.Mode != "all_for_one" && req.Mode != "friendship" {
		return c.Status(400).JSON(fiber.Map{"error": "Mode invalide"})
	}

	if req.Mode == "friendship" {
		pCountBody, _ := h.db.Select("challenge_participants",
			fmt.Sprintf("session_id=eq.%s&status=eq.accepted&select=id", req.SessionID), true)
		var pCountRows []map[string]interface{}
		json.Unmarshal(pCountBody, &pCountRows)
		if len(pCountRows) < 3 {
			return c.Status(400).JSON(fiber.Map{"error": "Le mode « Pouvoir de l'amitié » nécessite au moins 3 participants"})
		}
	}

	sBody, _ := h.db.Select("challenge_sessions",
		fmt.Sprintf("id=eq.%s&select=creator_id,status", req.SessionID), true)
	var sess []map[string]interface{}
	json.Unmarshal(sBody, &sess)
	if len(sess) == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Défi introuvable"})
	}
	if database.DBValue(sess[0]["creator_id"]) != user.ID {
		return c.Status(403).JSON(fiber.Map{"error": "Seul le créateur peut changer le mode"})
	}
	status, _ := sess[0]["status"].(string)
	if status != "waiting" {
		return c.Status(400).JSON(fiber.Map{"error": "Le mode ne peut être changé qu'en statut 'En attente'"})
	}

	h.db.Update("challenge_sessions", fmt.Sprintf("id=eq.%s", req.SessionID),
		[]byte(fmt.Sprintf(`{"reward_mode":"%s"}`, req.Mode)), true)
	return c.JSON(fiber.Map{"success": true})
}

func (h *Handler) distributeChallengeRewards(sessionID string) {
	sBody, _ := h.db.Select("challenge_sessions",
		fmt.Sprintf("id=eq.%s&select=status,total_xp_pool,reward_mode", sessionID), true)
	var sess []map[string]interface{}
	json.Unmarshal(sBody, &sess)
	if len(sess) == 0 {
		return
	}
	if sess[0]["status"] == "completed" {
		return
	}

	partBody, _ := h.db.Select("challenge_participants",
		fmt.Sprintf("session_id=eq.%s&status=eq.accepted&select=user_id", sessionID), true)
	var parts []map[string]interface{}
	json.Unmarshal(partBody, &parts)

	if len(parts) < 2 {
		return
	}

	scoreBody, _ := h.db.Select("challenge_scores",
		fmt.Sprintf("session_id=eq.%s&select=user_id,correct_count,time_taken_ms&order=correct_count.desc,time_taken_ms.asc", sessionID), true)
	var scores []map[string]interface{}
	json.Unmarshal(scoreBody, &scores)

	if len(scores) < 2 {
		return
	}

	totalPool := int(sess[0]["total_xp_pool"].(float64))
	rewardMode, _ := sess[0]["reward_mode"].(string)

	var rewards []int
	if rewardMode == "friendship" && len(scores) >= 3 {
		rewards = []int{50, 35, 15}
	} else {
		rewards = []int{100}
	}

	for i, s := range scores {
		if i >= len(rewards) {
			break
		}
		uid := database.DBValue(s["user_id"])
		pct := rewards[i]
		amount := totalPool * pct / 100
		if amount <= 0 {
			continue
		}
		txData, _ := json.Marshal(map[string]interface{}{
			"user_id":   uid,
			"source":    "challenge_reward",
			"source_id": sessionID,
			"amount":    amount,
		})
		h.db.Insert("xp_transactions", txData, true)

		pBody, _ := h.db.Select("user_profiles", fmt.Sprintf("id=eq.%s&select=xp", uid), true)
		var pRows []map[string]interface{}
		json.Unmarshal(pBody, &pRows)
		if len(pRows) > 0 {
			currentXP := int(pRows[0]["xp"].(float64))
			newXP := currentXP + amount
			newRank := getRankForXP(newXP, h.loadRanks())
			xpData, _ := json.Marshal(map[string]interface{}{"xp": newXP, "rank": newRank})
			h.db.Update("user_profiles", fmt.Sprintf("id=eq.%s", uid), xpData, true)
		}

		placeName := "er"
		if i == 1 {
			placeName = "ème"
		} else if i == 2 {
			placeName = "ème"
		}
		createNotification(h.db, uid, "challenge_reward",
			"🏆 Récompense défi",
			fmt.Sprintf("Tu as terminé %d%s avec %d XP gagnés !", i+1, placeName, amount),
			fmt.Sprintf(`{"session_id":"%s"}`, sessionID))
	}

	for i, s := range scores {
		uid := database.DBValue(s["user_id"])
		pBody, _ := h.db.Select("user_profiles", fmt.Sprintf("id=eq.%s&select=challenges_played,challenges_won", uid), true)
		var pRows []map[string]interface{}
		json.Unmarshal(pBody, &pRows)
		if len(pRows) > 0 {
			played := database.DBInt(pRows[0]["challenges_played"]) + 1
			won := database.DBInt(pRows[0]["challenges_won"])
			if i < len(rewards) {
				won++
			}
			statData, _ := json.Marshal(map[string]interface{}{"challenges_played": played, "challenges_won": won})
			h.db.Update("user_profiles", fmt.Sprintf("id=eq.%s", uid), statData, true)
		}
	}

	now := time.Now().UTC().Format(time.RFC3339)
	h.db.Update("challenge_sessions",
		fmt.Sprintf("id=eq.%s", sessionID),
		[]byte(fmt.Sprintf(`{"status":"completed","completed_at":"%s"}`, now)), true)
}

func (h *Handler) checkAndDistribute(sessionID string) {
	partBody, _ := h.db.Select("challenge_participants",
		fmt.Sprintf("session_id=eq.%s&status=eq.accepted&select=id", sessionID), true)
	var parts []map[string]interface{}
	json.Unmarshal(partBody, &parts)

	scoreBody, _ := h.db.Select("challenge_scores",
		fmt.Sprintf("session_id=eq.%s&select=id", sessionID), true)
	var scores []map[string]interface{}
	json.Unmarshal(scoreBody, &scores)

	if len(scores) >= len(parts) && len(parts) >= 2 {
		h.distributeChallengeRewards(sessionID)
	}
}
