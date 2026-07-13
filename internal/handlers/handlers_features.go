package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"otaku-quiz-africa/internal/database"
)

// ============================================================
// 0. RPC ENDPOINTS (Supabase RPC proxy)
// ============================================================

func (h *Handler) APIGetGlobalLeaderboard(c *fiber.Ctx) error {
	limitStr := c.Query("limit_count", "50")
	offsetStr := c.Query("offset_count", "0")
	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)
	if limit <= 0 {
		limit = 50
	}

	body, err := h.db.RPC("get_global_leaderboard", map[string]interface{}{
		"limit_count":  limit,
		"offset_count": offset,
	})
	if err != nil {
		return c.JSON([]interface{}{})
	}

	var result []map[string]interface{}
	json.Unmarshal(body, &result)
	return c.JSON(result)
}

func (h *Handler) APIGetMonthlyLeaderboard(c *fiber.Ctx) error {
	yearMonth := c.Query("year_month", "")
	limitStr := c.Query("limit_count", "50")
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 {
		limit = 50
	}

	params := map[string]interface{}{
		"limit_count": limit,
	}
	if yearMonth != "" {
		params["year_month"] = yearMonth
	}

	body, err := h.db.RPC("get_monthly_leaderboard", params)
	if err != nil {
		return c.JSON([]interface{}{})
	}

	var result []map[string]interface{}
	json.Unmarshal(body, &result)
	return c.JSON(result)
}

func (h *Handler) APIGetQuizLeaderboard(c *fiber.Ctx) error {
	quizID := c.Query("quiz_id", "")
	if quizID == "" {
		return c.JSON([]interface{}{})
	}

	body, err := h.db.RPC("get_quiz_leaderboard", map[string]interface{}{
		"quiz_id": quizID,
	})
	if err != nil {
		return c.JSON([]interface{}{})
	}

	var result []map[string]interface{}
	json.Unmarshal(body, &result)
	return c.JSON(result)
}

func (h *Handler) APIGetSeriesLeaderboard(c *fiber.Ctx) error {
	seriesName := c.Query("series_name", "")
	limitStr := c.Query("limit_count", "50")
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 {
		limit = 50
	}

	if seriesName == "" {
		return c.JSON([]interface{}{})
	}

	body, err := h.db.RPC("get_series_leaderboard", map[string]interface{}{
		"series_name": seriesName,
		"limit_count": limit,
	})
	if err != nil {
		return c.JSON([]interface{}{})
	}

	var result []map[string]interface{}
	json.Unmarshal(body, &result)
	return c.JSON(result)
}

// ============================================================
// 1. NOTIFICATIONS
// ============================================================

func (h *Handler) APIGetNotifications(c *fiber.Ctx) error {
	user := c.Locals("user").(*UserProfile)

	body, err := h.db.Select("notifications",
		fmt.Sprintf("user_id=eq.%s&order=created_at.desc&limit=50", user.ID), true)
	if err != nil {
		return c.JSON(fiber.Map{"notifications": []interface{}{}, "error": err.Error()})
	}
	var notifications []map[string]interface{}
	json.Unmarshal(body, &notifications)

	unreadBody, err := h.db.Select("notifications",
		fmt.Sprintf("user_id=eq.%s&is_read=eq.false&select=id", user.ID), true)
	unreadCount := 0
	if err == nil {
		var unread []interface{}
		json.Unmarshal(unreadBody, &unread)
		unreadCount = len(unread)
	}

	return c.JSON(fiber.Map{"notifications": notifications, "unread_count": unreadCount})
}

func (h *Handler) APIMarkNotificationsRead(c *fiber.Ctx) error {
	user := c.Locals("user").(*UserProfile)

	type Request struct {
		NotificationID string `json:"notification_id"`
		MarkAll        bool   `json:"mark_all"`
	}
	var req Request
	c.BodyParser(&req)

	if req.MarkAll {
		h.db.Update("notifications", fmt.Sprintf("user_id=eq.%s&is_read=eq.false", user.ID),
			[]byte(`{"is_read":true}`), true)
	} else if req.NotificationID != "" {
		h.db.Update("notifications", fmt.Sprintf("id=eq.%s&user_id=eq.%s", req.NotificationID, user.ID),
			[]byte(`{"is_read":true}`), true)
	}

	return c.JSON(fiber.Map{"success": true})
}

func createNotification(db interface{ Select(string, string, bool) ([]byte, error); Insert(string, []byte, bool) ([]byte, error) }, userID, nType, title, message string, extraData ...string) {
	extraJSON := "{}"
	if len(extraData) > 0 && extraData[0] != "" {
		extraJSON = extraData[0]
	}
	payload := map[string]interface{}{
		"user_id": userID,
		"type":    nType,
		"title":   title,
		"message": message,
		"data":    json.RawMessage(extraJSON),
	}
	data, _ := json.Marshal(payload)
	db.Insert("notifications", data, true)
}

// ============================================================
// ANNOUNCEMENTS API
// ============================================================

func (h *Handler) APICreateAnnouncement(c *fiber.Ctx) error {
	user := c.Locals("user").(*UserProfile)
	if !user.IsAdmin {
		return c.Status(403).JSON(fiber.Map{"error": "Non autorisé"})
	}

	type Request struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		ImageURL    string `json:"image_url"`
		QuizID      string `json:"quiz_id"`
		Type        string `json:"type"`
		EndsAt      string `json:"ends_at"`
	}
	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Données invalides"})
	}
	if req.Title == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Le titre est requis"})
	}
	if req.Type == "" {
		req.Type = "quiz"
	}

	payload := map[string]interface{}{
		"title":       req.Title,
		"description": req.Description,
		"image_url":   req.ImageURL,
		"type":        req.Type,
		"status":      "active",
		"created_by":  user.ID,
	}
	if req.QuizID != "" {
		payload["quiz_id"] = req.QuizID
	}
	if req.EndsAt != "" {
		payload["ends_at"] = req.EndsAt
	}

	jsonData, _ := json.Marshal(payload)
	body, err := h.db.Insert("announcements", jsonData, true)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Erreur création annonce"})
	}

	var result []map[string]interface{}
	json.Unmarshal(body, &result)
	id := ""
	if len(result) > 0 {
		if uid, ok := result[0]["id"]; ok {
			id = fmt.Sprintf("%v", uid)
		}
	}
	return c.JSON(fiber.Map{"success": true, "id": id})
}

func (h *Handler) APIDeleteAnnouncement(c *fiber.Ctx) error {
	user := c.Locals("user").(*UserProfile)
	if !user.IsAdmin {
		return c.Status(403).JSON(fiber.Map{"error": "Non autorisé"})
	}

	type Request struct {
		ID string `json:"id"`
	}
	var req Request
	if err := c.BodyParser(&req); err != nil || req.ID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "ID requis"})
	}

	err := h.db.Delete("announcements", fmt.Sprintf("id=eq.%s", req.ID), true)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Erreur suppression"})
	}
	return c.JSON(fiber.Map{"success": true})
}

func (h *Handler) APIUpdateAnnouncement(c *fiber.Ctx) error {
	user := c.Locals("user").(*UserProfile)
	if !user.IsAdmin {
		return c.Status(403).JSON(fiber.Map{"error": "Non autorisé"})
	}

	type Request struct {
		ID          string `json:"id"`
		Title       string `json:"title"`
		Description string `json:"description"`
		ImageURL    string `json:"image_url"`
		Type        string `json:"type"`
		Status      string `json:"status"`
		EndsAt      string `json:"ends_at"`
	}
	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Données invalides"})
	}
	if req.ID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "ID requis"})
	}

	updates := map[string]interface{}{}
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Description != "" {
		updates["description"] = req.Description
	} else {
		updates["description"] = nil
	}
	if req.ImageURL != "" {
		updates["image_url"] = req.ImageURL
	} else {
		updates["image_url"] = nil
	}
	if req.Type != "" {
		updates["type"] = req.Type
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}
	if req.EndsAt != "" {
		updates["ends_at"] = req.EndsAt
	} else {
		updates["ends_at"] = nil
	}

	jsonData, _ := json.Marshal(updates)
	_, err := h.db.Update("announcements", fmt.Sprintf("id=eq.%s", req.ID), jsonData, true)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Erreur mise à jour"})
	}
	return c.JSON(fiber.Map{"success": true})
}

func (h *Handler) APIGetAllAnnouncements(c *fiber.Ctx) error {
	user := c.Locals("user").(*UserProfile)
	if !user.IsAdmin {
		return c.Status(403).JSON(fiber.Map{"error": "Non autorisé"})
	}

	body, err := h.db.Select("announcements", "order=created_at.desc", true)
	if err != nil {
		return c.JSON([]interface{}{})
	}
	var announcements []map[string]interface{}
	json.Unmarshal(body, &announcements)
	return c.JSON(announcements)
}

// ============================================================
// OFFICIAL QUIZZES API
// ============================================================

func (h *Handler) APICreateOfficialQuiz(c *fiber.Ctx) error {
	user := c.Locals("user").(*UserProfile)
	if !user.IsAdmin {
		return c.Status(403).JSON(fiber.Map{"error": "Non autorisé"})
	}

	type Reward struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		URL         string `json:"url"`
		RankFrom    int    `json:"rank_from"`
		RankTo      int    `json:"rank_to"`
	}
	type Request struct {
		Title           string   `json:"title"`
		Description     string   `json:"description"`
		Category        string   `json:"category"`
		Subcategory     string   `json:"subcategory"`
		Series          string   `json:"series"`
		StartsAt        string   `json:"starts_at"`
		EndsAt          string   `json:"ends_at"`
		DurationSeconds int      `json:"duration_seconds"`
		DurationMode    string   `json:"duration_mode"`
		Rewards         []Reward `json:"rewards"`
	}
	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Données invalides"})
	}
	if req.Title == "" || req.Series == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Titre et série requis"})
	}
	if req.Category == "" {
		req.Category = "Shonen"
	}
	if req.Subcategory == "" {
		req.Subcategory = "Général"
	}
	if req.DurationSeconds <= 0 {
		req.DurationSeconds = 30
	}
	if req.DurationMode == "" {
		req.DurationMode = "per_question"
	}

	status := "active"
	if req.StartsAt != "" {
		startsAt, err := time.Parse("2006-01-02T15:04", req.StartsAt)
		if err == nil && startsAt.After(time.Now()) {
			status = "scheduled"
		}
	}

	quizPayload := map[string]interface{}{
		"creator_id":         user.ID,
		"title":              req.Title,
		"description":        req.Description,
		"category":           req.Category,
		"subcategory":        req.Subcategory,
		"series":             req.Series,
		"quiz_type":          "official",
		"status":             status,
		"starts_at":          req.StartsAt,
		"ends_at":            req.EndsAt,
		"duration_seconds":   req.DurationSeconds,
		"duration_mode":      req.DurationMode,
		"leaderboard_public": true,
		"is_visible":         true,
	}

	jsonData, _ := json.Marshal(quizPayload)
	body, err := h.db.Insert("quizzes", jsonData, true)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Erreur création quiz"})
	}

	var result []map[string]interface{}
	json.Unmarshal(body, &result)
	quizID := ""
	if len(result) > 0 {
		if id, ok := result[0]["id"]; ok {
			quizID = fmt.Sprintf("%v", id)
		}
	}

	for _, r := range req.Rewards {
		rewardPayload := map[string]interface{}{
			"quiz_id":     quizID,
			"title":       r.Title,
			"description": r.Description,
			"url":         r.URL,
			"rank_from":   r.RankFrom,
			"rank_to":     r.RankTo,
		}
		rewardJSON, _ := json.Marshal(rewardPayload)
		h.db.Insert("quiz_rewards", rewardJSON, true)
	}

	return c.JSON(fiber.Map{"success": true, "id": quizID})
}

func (h *Handler) APIGetAllOfficialQuizzes(c *fiber.Ctx) error {
	user := c.Locals("user").(*UserProfile)
	if !user.IsAdmin {
		return c.Status(403).JSON(fiber.Map{"error": "Non autorisé"})
	}

	body, err := h.db.Select("quizzes", "quiz_type=eq.official&order=created_at.desc&select=*", true)
	if err != nil {
		return c.JSON([]interface{}{})
	}
	var quizzes []map[string]interface{}
	json.Unmarshal(body, &quizzes)
	return c.JSON(quizzes)
}

func (h *Handler) NotificationsPage(c *fiber.Ctx) error {
	return renderPage(c, "Notifications", `
<div class="notif-page">
    <h1 class="page-title">🔔 Notifications</h1>
    <div class="notif-actions">
        <button class="btn-sm btn-outline" onclick="markAllRead()">Tout marquer comme lu</button>
    </div>
    <div id="notifications-list" class="notif-list">
        <div class="fr-loading">Chargement...</div>
    </div>
</div>
<style>
.notif-page{max-width:600px;margin:0 auto}
.notif-actions{margin-bottom:16px;text-align:right}
.notif-list{display:flex;flex-direction:column;gap:8px}
.notif-item{background:#16213e;border:1px solid #2d2d44;border-radius:8px;padding:16px;display:flex;gap:12px;align-items:flex-start;transition:all .2s}
.notif-item.unread{border-left:3px solid #6366f1}
.notif-item.read{opacity:.7}
.notif-icon{font-size:1.5rem;min-width:32px;text-align:center}
.notif-content{flex:1}
.notif-title{font-weight:600;font-size:.9rem}
.notif-message{color:#94a3b8;font-size:.8rem;margin-top:4px}
.notif-time{color:#64748b;font-size:.7rem;margin-top:4px}
.notif-empty{text-align:center;padding:40px;color:#94a3b8}
</style>
<script>
(function(){
    loadNotifications();
    setInterval(loadNotifications,10000);
})();
function loadNotifications(){
    fetch('/api/notifications').then(function(r){return r.json()}).then(function(d){
        var el=document.getElementById('notifications-list');
        if(!d.notifications||d.notifications.length===0){
            el.innerHTML='<div class="notif-empty">Aucune notification</div>';return;
        }
        var h='';d.notifications.forEach(function(n){
            var icon='🔔';if(n.type==='friend_request')icon='👋';else if(n.type==='badge_unlocked')icon='🏅';else if(n.type==='quiz_completed')icon='🎮';else if(n.type==='event_starting')icon='📅';else if(n.type==='rank_up')icon='⬆️';else if(n.type==='challenge_invitation')icon='⚔️';
            var cls=n.is_read?'read':'unread';
            h+='<div class="notif-item '+cls+'" onclick="handleNotifClick(\''+n.id+'\',\''+n.type+'\','+(n.data?JSON.stringify(n.data).replace(/'/g,"\\'"):'null')+')"><div class="notif-icon">'+icon+'</div><div class="notif-content"><div class="notif-title">'+n.title+'</div><div class="notif-message">'+n.message+'</div><div class="notif-time">'+new Date(n.created_at).toLocaleString('fr-FR')+'</div></div></div>';
        });
        el.innerHTML=h;
        var badge=document.querySelector('.notif-badge-count');
        if(badge){badge.textContent=d.unread_count||'';badge.style.display=d.unread_count>0?'inline':'none';}
    });
}
function markRead(id){fetch('/api/notifications/read',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({notification_id:id})}).then(function(){loadNotifications()});}
function markAllRead(){fetch('/api/notifications/read',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({mark_all:true})}).then(function(){loadNotifications()});}
function handleNotifClick(id,type,data){markRead(id);if(type==='challenge_invitation'&&data&&data.session_id){window.location.href='/challenges/'+data.session_id;}}
</script>
	`)
}

// ============================================================
// 2. STREAK SYSTEM
// ============================================================

func (h *Handler) updateStreak(userID string) {
	today := time.Now().Format("2006-01-02")

	body, err := h.db.Select("user_profiles",
		fmt.Sprintf("id=eq.%s&select=current_streak,longest_streak,last_login_date", userID), true)
	if err != nil {
		return
	}
	var profiles []map[string]interface{}
	json.Unmarshal(body, &profiles)
	if len(profiles) == 0 {
		return
	}

	lastLogin, _ := profiles[0]["last_login_date"].(string)
	currentStreak, _ := profiles[0]["current_streak"].(float64)
	longestStreak, _ := profiles[0]["longest_streak"].(float64)

	newStreak := int(currentStreak)
	if lastLogin == today {
		h.db.Update("user_profiles", fmt.Sprintf("id=eq.%s", userID),
			[]byte(fmt.Sprintf(`{"last_login_date":"%s"}`, today)), true)
		return
	}

	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	if lastLogin == yesterday {
		newStreak++
	} else {
		newStreak = 1
	}

	newLongest := int(longestStreak)
	if newStreak > newLongest {
		newLongest = newStreak
	}

	updateData := fmt.Sprintf(`{"current_streak":%d,"longest_streak":%d,"last_login_date":"%s"}`, newStreak, newLongest, today)
	h.db.Update("user_profiles", fmt.Sprintf("id=eq.%s", userID), []byte(updateData), true)
}

// ============================================================
// 3. ANNOUNCEMENTS FROM DB
// ============================================================

func (h *Handler) Announcements(c *fiber.Ctx) error {
	body, err := h.db.Select("announcements",
		"status=eq.active&order=starts_at.desc&limit=20", true)
	if err != nil {
		return renderPage(c, "Annonces", `<div class="empty-state"><p>Aucune annonce pour le moment</p></div>`)
	}
	var announcements []map[string]interface{}
	json.Unmarshal(body, &announcements)

	items := ""
	for _, a := range announcements {
		title, _ := a["title"].(string)
		desc, _ := a["description"].(string)
		imgURL, _ := a["image_url"].(string)
		aType, _ := a["type"].(string)
		icon := "📢"
		if aType == "event" {
			icon = "🎯"
		} else if aType == "news" {
			icon = "📰"
		}

		imgHTML := ""
		if imgURL != "" {
			imgHTML = fmt.Sprintf(`<img src="%s" alt="" class="ann-img">`, imgURL)
		}

		quizID, _ := a["quiz_id"].(string)
		linkHTML := ""
		if quizID != "" {
			linkHTML = fmt.Sprintf(`<a href="/quiz/%s" class="btn-primary btn-sm" style="margin-top:12px;display:inline-block">Jouer au quiz</a>`, quizID)
		}

		items += fmt.Sprintf(`
<div class="ann-card">
    <div class="ann-icon">%s</div>
    <div class="ann-content">
        <h3 class="ann-title">%s</h3>
        <p class="ann-desc">%s</p>
        %s%s
    </div>
</div>`, icon, title, desc, imgHTML, linkHTML)
	}

	if items == "" {
		items = `<div class="empty-state"><p>Aucune annonce pour le moment</p></div>`
	}

	return renderPage(c, "Annonces", fmt.Sprintf(`
<div class="ann-page">
    <h1 class="page-title">📢 Annonces</h1>
    %s
</div>
<style>
.ann-page{max-width:600px;margin:0 auto}
.ann-card{background:#16213e;border:1px solid #2d2d44;border-radius:12px;padding:20px;margin-bottom:12px;display:flex;gap:16px;align-items:flex-start}
.ann-icon{font-size:2rem;min-width:40px;text-align:center}
.ann-content{flex:1}
.ann-title{font-weight:600;margin-bottom:8px}
.ann-desc{color:#94a3b8;font-size:.9rem}
.ann-img{width:100%%;border-radius:8px;margin-top:12px;max-height:200px;object-fit:cover}
.empty-state{text-align:center;padding:60px 20px;color:#94a3b8}
</style>
	`, items))
}

// ============================================================
// 4. FAQ FROM DB
// ============================================================

func (h *Handler) FAQFromDB(c *fiber.Ctx) error {
	body, err := h.db.Select("faq_entries",
		"visible=eq.true&order=order_index.asc", true)

	sections := ""
	if err == nil {
		var entries []map[string]interface{}
		json.Unmarshal(body, &entries)

		themeMap := map[string][]map[string]interface{}{}
		themeOrder := []string{}
		for _, e := range entries {
			theme, _ := e["theme"].(string)
			if _, ok := themeMap[theme]; !ok {
				themeMap[theme] = []map[string]interface{}{}
				themeOrder = append(themeOrder, theme)
			}
			themeMap[theme] = append(themeMap[theme], e)
		}

		for _, theme := range themeOrder {
			items := themeMap[theme]
			questions := ""
			for _, item := range items {
				q, _ := item["question"].(string)
				a, _ := item["answer"].(string)
				questions += fmt.Sprintf(`
<div class="faq-item">
    <strong>%s</strong>
    <p>%s</p>
</div>`, q, a)
			}
			sections += fmt.Sprintf(`
<div class="faq-section">
    <h2>%s</h2>
    %s
</div>`, theme, questions)
		}
	}

	if sections == "" {
		sections = `<div class="empty-state"><p>Aucune question pour le moment</p></div>`
	}

	return renderPage(c, "FAQ", fmt.Sprintf(`
<h1 style="margin-bottom: 24px;">❓ FAQ</h1>
%s
<style>
.faq-section{margin-bottom:32px}
.faq-section h2{font-size:1.25rem;margin-bottom:16px;color:#6366f1}
.faq-item{background:#16213e;border:1px solid #2d2d44;border-radius:8px;padding:16px;margin-bottom:12px}
.faq-item strong{display:block;margin-bottom:8px}
.faq-item p{color:#94a3b8;font-size:0.9rem}
.empty-state{text-align:center;padding:60px 20px;color:#94a3b8}
</style>
	`, sections))
}

// ============================================================
// 5. ADMIN CHAT
// ============================================================

func (h *Handler) APIGetAdminConversations(c *fiber.Ctx) error {
	user := c.Locals("user").(*UserProfile)

	body, err := h.db.Select("admin_conversations",
		fmt.Sprintf("user_id=eq.%s&order=last_message_at.desc", user.ID), true)
	if err != nil {
		return c.JSON(fiber.Map{"conversations": []interface{}{}, "error": err.Error()})
	}
	var conversations []map[string]interface{}
	json.Unmarshal(body, &conversations)
	return c.JSON(fiber.Map{"conversations": conversations})
}

func (h *Handler) APICreateAdminConversation(c *fiber.Ctx) error {
	user := c.Locals("user").(*UserProfile)

	type Request struct {
		Subject string `json:"subject"`
	}
	var req Request
	c.BodyParser(&req)

	if req.Subject == "" {
		req.Subject = "Support"
	}

	insertData := fmt.Sprintf(`{"user_id":"%s","subject":"%s"}`, user.ID, req.Subject)
	insertBody, err := h.db.Insert("admin_conversations", []byte(insertData), true)
	if err != nil {
		return c.JSON(fiber.Map{"error": err.Error()})
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

func (h *Handler) APIGetAdminMessages(c *fiber.Ctx) error {
	convID := c.Query("conversation_id")

	body, err := h.db.Select("admin_messages",
		fmt.Sprintf("conversation_id=eq.%s&order=created_at.asc&limit=100", convID), true)
	if err != nil {
		return c.JSON(fiber.Map{"messages": []interface{}{}, "error": err.Error()})
	}
	var messages []map[string]interface{}
	json.Unmarshal(body, &messages)
	return c.JSON(fiber.Map{"messages": messages})
}

func (h *Handler) APISendAdminMessage(c *fiber.Ctx) error {
	user := c.Locals("user").(*UserProfile)

	type Request struct {
		ConversationID string `json:"conversation_id"`
		Content        string `json:"content"`
	}
	var req Request
	c.BodyParser(&req)

	insertData, _ := json.Marshal(map[string]string{
		"conversation_id": req.ConversationID,
		"sender_id":       user.ID,
		"content":         req.Content,
	})
	_, err := h.db.Insert("admin_messages", insertData, true)
	if err != nil {
		return c.JSON(fiber.Map{"error": err.Error()})
	}

	h.db.Update("admin_conversations",
		fmt.Sprintf("id=eq.%s", req.ConversationID),
		[]byte(`{"status":"open"}`), true)

	return c.JSON(fiber.Map{"success": true})
}

func (h *Handler) APIDeleteAdminConversation(c *fiber.Ctx) error {
	user := c.Locals("user").(*UserProfile)

	type Request struct {
		ConversationID string `json:"conversation_id"`
	}
	var req Request
	if err := c.BodyParser(&req); err != nil || req.ConversationID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "ID requis"})
	}

	convBody, err := h.db.Select("admin_conversations",
		fmt.Sprintf("id=eq.%s&select=user_id", req.ConversationID), true)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Erreur"})
	}
	var convs []map[string]interface{}
	json.Unmarshal(convBody, &convs)
	if len(convs) == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Ticket introuvable"})
	}

	convUserID, _ := convs[0]["user_id"].(string)
	if convUserID != user.ID && !user.IsAdmin {
		return c.Status(403).JSON(fiber.Map{"error": "Non autorisé"})
	}

	h.db.Delete("admin_messages", fmt.Sprintf("conversation_id=eq.%s", req.ConversationID), true)
	h.db.Delete("admin_conversations", fmt.Sprintf("id=eq.%s", req.ConversationID), true)

	return c.JSON(fiber.Map{"success": true})
}

func (h *Handler) APIUpdateAdminConversationStatus(c *fiber.Ctx) error {
	type Request struct {
		ConversationID string `json:"conversation_id"`
		Status         string `json:"status"`
	}
	var req Request
	if err := c.BodyParser(&req); err != nil || req.ConversationID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "ID requis"})
	}
	if req.Status != "open" && req.Status != "assigned" && req.Status != "closed" {
		return c.Status(400).JSON(fiber.Map{"error": "Statut invalide"})
	}

	data, _ := json.Marshal(map[string]string{"status": req.Status})
	_, err := h.db.Update("admin_conversations",
		fmt.Sprintf("id=eq.%s", req.ConversationID), data, true)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Erreur"})
	}
	return c.JSON(fiber.Map{"success": true})
}

func (h *Handler) SupportPage(c *fiber.Ctx) error {
	return renderPage(c, "Support", `
<div class="support-page">
    <h1 class="page-title">💬 Support</h1>
    <p class="text-muted" style="margin-bottom:24px">Besoin d'aide ? Décrivez votre problème et envoyez un message.</p>

    <div id="new-ticket-form" style="margin-bottom:24px">
        <div class="card" style="padding:16px">
            <div style="font-weight:600;margin-bottom:8px">Nouveau ticket</div>
            <input type="text" id="ticket-subject" placeholder="Sujet (optionnel)" style="width:100%;padding:10px;border-radius:8px;border:1px solid #2d2d44;background:#0f172a;color:white;margin-bottom:8px">
            <div style="display:flex;gap:8px">
                <input type="text" id="ticket-message" placeholder="Votre message..." autocomplete="off" style="flex:1;padding:10px;border-radius:8px;border:1px solid #2d2d44;background:#0f172a;color:white">
                <button class="btn-primary btn-sm" onclick="createAndSend()">Envoyer</button>
            </div>
        </div>
    </div>

    <h2 style="font-size:1rem;margin-bottom:12px;color:#94a3b8">Mes tickets</h2>
    <div id="support-conversations" class="fr-list"></div>
</div>
<script>
function loadSupportConvos(){
    fetch('/api/admin/conversations').then(function(r){return r.json()}).then(function(d){
        var el=document.getElementById('support-conversations');
        if(!d.conversations||d.conversations.length===0){el.innerHTML='<p class="text-muted">Aucun ticket</p>';return;}
        var h='';d.conversations.forEach(function(c){
            var st=c.status==='open'?'Ouvert':c.status==='assigned'?'Assigné':'Fermé';
            var bc=c.status==='open'?'badge-primary':c.status==='assigned'?'badge-success':'badge-outline';
            h+='<div class="fr-card" style="cursor:default"><div class="fr-info" onclick="window.location.href=\'/support/'+c.id+'\'" style="cursor:pointer;flex:1"><div class="fr-name">'+c.subject+'</div><div class="fr-meta text-sm text-muted">'+new Date(c.created_at).toLocaleString('fr-FR')+'</div></div><div class="fr-actions"><span class="'+bc+'">'+st+'</span><button class="btn-ghost btn-sm" style="color:#ef4444;padding:4px 8px" onclick="event.stopPropagation();deleteTicket(\''+c.id+'\')">🗑</button></div></div>';
        });
        el.innerHTML=h;
    });
}
function createAndSend(){
    var msg=document.getElementById('ticket-message').value.trim();
    if(!msg){alert('Veuillez écrire un message');return;}
    var subject=document.getElementById('ticket-subject').value.trim()||'Support';
    fetch('/api/admin/conversations/create',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({subject:subject})})
    .then(function(r){return r.json()}).then(function(d){
        if(!d.conversation_id){alert('Erreur création ticket');return;}
        fetch('/api/admin/messages/send',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({conversation_id:d.conversation_id,content:msg})})
        .then(function(r){return r.json()}).then(function(){
            window.location.href='/support/'+d.conversation_id;
        });
    });
}
document.getElementById('ticket-message').addEventListener('keydown',function(e){
    if(e.key==='Enter'&&!e.shiftKey){e.preventDefault();createAndSend();}
});
function deleteTicket(id){
    if(!confirm('Supprimer ce ticket et tous ses messages ?'))return;
    fetch('/api/admin/conversations/delete',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({conversation_id:id})})
    .then(function(r){return r.json()}).then(function(d){
        if(d.error){alert(d.error);return;}
        loadSupportConvos();
    });
}
loadSupportConvos();
</script>
<style>
.support-page{max-width:600px;margin:0 auto}
</style>
	`)
}

func (h *Handler) SupportConversation(c *fiber.Ctx) error {
	user := c.Locals("user").(*UserProfile)
	convID := c.Params("id")

	convBody, err := h.db.Select("admin_conversations",
		fmt.Sprintf("id=eq.%s&user_id=eq.%s&select=*", convID, user.ID), true)
	if err != nil {
		return c.Redirect("/support")
	}
	var convs []map[string]interface{}
	json.Unmarshal(convBody, &convs)
	if len(convs) == 0 {
		return c.Redirect("/support")
	}
	conv := convs[0]
	subject, _ := conv["subject"].(string)
	status, _ := conv["status"].(string)
	statusLabel := "Ouvert"
	if status == "assigned" {
		statusLabel = "Assigné"
	} else if status == "closed" {
		statusLabel = "Fermé"
	}

	return renderPage(c, "Support - "+subject, fmt.Sprintf(`
<a href="/support" style="color:#6366f1;display:inline-block;margin-bottom:16px">← Retour aux tickets</a>
<div class="support-conv">
    <div class="conv-header">
        <h1 style="font-size:1.3rem;margin:0">%s</h1>
        <div style="display:flex;align-items:center;gap:10px">
            <span class="badge-outline">%s</span>
            <button class="btn-ghost btn-sm" style="color:#ef4444" onclick="deleteTicket()">🗑 Supprimer</button>
        </div>
    </div>
    <div id="messages-list" class="messages-list">
        <div class="msg-system">Chargement...</div>
    </div>
    <form id="msg-form" class="msg-form">
        <input type="text" id="msg-input" placeholder="Écrire un message..." autocomplete="off" style="flex:1;padding:10px;border-radius:8px;border:1px solid #2d2d44;background:#0f172a;color:white">
        <button type="submit" class="btn-primary btn-sm">Envoyer</button>
    </form>
</div>
<style>
.support-conv{max-width:700px;margin:0 auto}
.conv-header{display:flex;justify-content:space-between;align-items:center;margin-bottom:16px;padding:12px 16px;background:#16213e;border-radius:8px;border:1px solid #2d2d44}
.messages-list{display:flex;flex-direction:column;gap:4px;max-height:60vh;overflow-y:auto;padding:12px;background:#0b141a;border:1px solid #2d2d44;border-radius:8px;margin-bottom:12px}
.msg-bubble{max-width:72%%;padding:8px 12px 6px;border-radius:10px;font-size:0.88rem;line-height:1.4;position:relative;margin-bottom:2px}
.msg-me{align-self:flex-end;background:#005c4b;border-bottom-right-radius:4px}
.msg-them{align-self:flex-start;background:#1e2a38;border-bottom-left-radius:4px}
.msg-text{word-wrap:break-word;margin-bottom:2px}
.msg-meta{display:flex;align-items:center;justify-content:flex-end;gap:4px;margin-top:0}
.msg-time{font-size:0.65rem;color:rgba(255,255,255,0.45)}
.msg-check{font-size:0.7rem;color:#53bdeb}
.msg-system{text-align:center;color:#94a3b8;font-size:0.78rem;padding:8px 16px;background:rgba(255,255,255,0.04);border-radius:8px;margin:4px 0;align-self:center}
.msg-form{display:flex;gap:8px}
#messages-list{scroll-behavior:smooth}
</style>
<script>
var convId='%s';
function loadMessages(){
    fetch('/api/admin/messages?conversation_id='+convId).then(function(r){return r.json()}).then(function(d){
        var el=document.getElementById('messages-list');
        if(!d.messages||d.messages.length===0){el.innerHTML='<div class="msg-system">Aucun message. Commencez la conversation !</div>';return;}
        var h='';d.messages.forEach(function(m){
            var isSystem=m.sender_id==='system';
            if(isSystem){
                h+='<div class="msg-system">'+escapeMsg(m.content)+'</div>';
            } else {
                var isMe=m.sender_id==='%s';
                var cls=isMe?'msg-me':'msg-them';
                var time=m.created_at?fmtTime(new Date(m.created_at)):'';
                var check=isMe?'<span class="msg-check">\u2713\u2713</span>':'';
                h+='<div class="msg-bubble '+cls+'"><div class="msg-text">'+escapeMsg(m.content)+'</div><div class="msg-meta"><span class="msg-time">'+time+'</span>'+check+'</div></div>';
            }
        });
        el.innerHTML=h;
        el.scrollTop=el.scrollHeight;
    });
}
function fmtTime(d){var h=d.getHours();var m=d.getMinutes();return(h<10?'0':'')+h+':'+(m<10?'0':'')+m;}
function escapeMsg(t){if(!t)return'';var d=document.createElement('div');d.appendChild(document.createTextNode(t));return d.innerHTML;}
document.getElementById('msg-form').addEventListener('submit',function(e){
    e.preventDefault();
    var input=document.getElementById('msg-input');
    var content=input.value.trim();
    if(!content)return;
    input.value='';
    fetch('/api/admin/messages/send',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({conversation_id:convId,content:content})})
    .then(function(r){return r.json()}).then(function(){loadMessages();});
});
loadMessages();
setInterval(loadMessages,5000);
function deleteTicket(){
    if(!confirm('Supprimer ce ticket et tous ses messages ?'))return;
    fetch('/api/admin/conversations/delete',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({conversation_id:convId})})
    .then(function(r){return r.json()}).then(function(d){
        if(d.error){alert(d.error);return;}
        window.location.href='/support';
    });
}
</script>
	`, subject, statusLabel, convID, user.ID))
}

// ============================================================
// 6. XP HISTORY
// ============================================================

func (h *Handler) APIGetXPHistory(c *fiber.Ctx) error {
	user := c.Locals("user").(*UserProfile)

	body, err := h.db.Select("xp_transactions",
		fmt.Sprintf("user_id=eq.%s&order=created_at.desc&limit=50", user.ID), true)
	if err != nil {
		return c.JSON(fiber.Map{"transactions": []interface{}{}, "error": err.Error()})
	}
	var transactions []map[string]interface{}
	json.Unmarshal(body, &transactions)
	return c.JSON(fiber.Map{"transactions": transactions})
}

func (h *Handler) XPHistoryPage(c *fiber.Ctx) error {
	user := c.Locals("user").(*UserProfile)
	return renderPage(c, "Historique XP", fmt.Sprintf(`
<div class="xp-page">
    <h1 class="page-title">💰 Historique XP</h1>
    <div class="xp-summary">
        <div class="ds-card ds-brand"><div class="ds-value">%d</div><div class="ds-label">XP Total</div></div>
    </div>
    <div id="xp-list" class="notif-list"><div class="fr-loading">Chargement...</div></div>
</div>
<script>
fetch('/api/xp/history').then(function(r){return r.json()}).then(function(d){
    var el=document.getElementById('xp-list');
    if(!d.transactions||d.transactions.length===0){el.innerHTML='<p class="text-muted">Aucune transaction</p>';return;}
    var h='';d.transactions.forEach(function(t){
        var icon='🎮';if(t.source==='streak')icon='🔥';else if(t.source==='challenge')icon='⚔️';else if(t.source==='event')icon='📅';
        var sign=t.amount>0?'+':'';
        var cls=t.amount>0?'color:#22c55e':'color:#ef4444';
        h+='<div class="notif-item"><div class="notif-icon">'+icon+'</div><div class="notif-content"><div class="notif-title">'+t.source+'</div><div class="notif-time">'+new Date(t.created_at).toLocaleString('fr-FR')+'</div></div><div style="font-weight:bold;font-size:1.1rem;'+cls+'">'+sign+Math.round(t.amount)+' XP</div></div>';
    });
    el.innerHTML=h;
});
</script>
<style>
.xp-page{max-width:600px;margin:0 auto}
.xp-summary{margin-bottom:24px;display:flex;gap:12px}
</style>
	`, user.XP))
}

// ============================================================
// 7. QUIZ REWARDS
// ============================================================

func (h *Handler) getQuizRewards(quizID string) []map[string]interface{} {
	body, err := h.db.Select("quiz_rewards",
		fmt.Sprintf("quiz_id=eq.%s&order=rank_from.asc", quizID), true)
	if err != nil {
		return nil
	}
	var rewards []map[string]interface{}
	json.Unmarshal(body, &rewards)
	return rewards
}

func renderQuizRewards(rewards []map[string]interface{}) string {
	if len(rewards) == 0 {
		return ""
	}

	html := `<div class="qr-section"><h3>🏆 Récompenses</h3>`
	for _, r := range rewards {
		title, _ := r["title"].(string)
		desc, _ := r["description"].(string)
		rankFrom, _ := r["rank_from"].(float64)
		rankTo, _ := r["rank_to"].(float64)
		url, _ := r["url"].(string)

		rangeText := fmt.Sprintf("#%d", int(rankFrom))
		if int(rankFrom) != int(rankTo) {
			rangeText = fmt.Sprintf("#%d - #%d", int(rankFrom), int(rankTo))
		}

		linkHTML := ""
		if url != "" {
			linkHTML = fmt.Sprintf(`<a href="%s" target="_blank" class="btn-outline btn-sm" style="margin-top:8px;display:inline-block">Voir la récompense</a>`, url)
		}

		html += fmt.Sprintf(`
<div class="qr-card">
    <div class="qr-rank">%s</div>
    <div class="qr-info">
        <div class="qr-title">%s</div>
        <div class="qr-desc text-muted text-sm">%s</div>
        %s
    </div>
</div>`, rangeText, title, desc, linkHTML)
	}
	html += `</div>`
	return html
}

// ============================================================
// 8. RANK CONFIG FROM DB
// ============================================================

type RankInfo struct {
	Label string `json:"label"`
	XP    int    `json:"xp_required"`
	Order int    `json:"display_order"`
}

var cachedRanks []RankInfo

func (h *Handler) loadRanks() []RankInfo {
	if len(cachedRanks) > 0 {
		return cachedRanks
	}
	body, err := h.db.Select("rank_config", "order=display_order.asc", true)
	if err != nil {
		return defaultRanks()
	}
	var ranks []map[string]interface{}
	json.Unmarshal(body, &ranks)

	result := []RankInfo{}
	for _, r := range ranks {
		label, _ := r["rank_label"].(string)
		xp, _ := r["xp_required"].(float64)
		order, _ := r["display_order"].(float64)
		result = append(result, RankInfo{Label: label, XP: int(xp), Order: int(order)})
	}
	if len(result) == 0 {
		return defaultRanks()
	}
	cachedRanks = result
	return result
}

func defaultRanks() []RankInfo {
	return []RankInfo{
		{"F", 0, 1}, {"E", 100, 2}, {"D", 500, 3}, {"C", 1500, 4},
		{"B", 3000, 5}, {"A", 6000, 6}, {"S", 10000, 7}, {"SS", 15000, 8},
		{"Z", 25000, 9}, {"Hunter", 40000, 10}, {"Nation", 60000, 11},
		{"Kage", 80000, 12}, {"OUTERVERSAL", 100000, 13},
	}
}

func getRankForXP(xp int, ranks []RankInfo) string {
	rank := "F"
	for _, r := range ranks {
		if xp >= r.XP {
			rank = r.Label
		}
	}
	return rank
}

func getXPForNextRank(xp int, ranks []RankInfo) (int, string) {
	for _, r := range ranks {
		if xp < r.XP {
			return r.XP - xp, r.Label
		}
	}
	return 0, "MAX"
}

func calcXPPercentFromRanks(xp int, ranks []RankInfo) int {
	prevXP := 0
	for _, r := range ranks {
		if xp < r.XP {
			if r.XP == prevXP {
				return 100
			}
			return int(math.Min(100, float64(xp-prevXP)/float64(r.XP-prevXP)*100))
		}
		prevXP = r.XP
	}
	return 100
}

// ============================================================
// 9. USER QUIZ ATTEMPTS
// ============================================================

func (h *Handler) trackQuizAttempt(userID, quizID string, score int, xpEarned int) {
	attemptBody, err := h.db.Select("user_quiz_attempts",
		fmt.Sprintf("user_id=eq.%s&quiz_id=eq.%s&order=attempt_number.desc&limit=1", userID, quizID), true)
	if err != nil {
		return
	}
	var attempts []map[string]interface{}
	json.Unmarshal(attemptBody, &attempts)

	attemptNum := 1
	if len(attempts) > 0 {
		if an, ok := attempts[0]["attempt_number"].(float64); ok {
			attemptNum = int(an) + 1
		}
	}

	insertData := fmt.Sprintf(`{"user_id":"%s","quiz_id":"%s","attempt_number":%d,"score":%d,"xp_earned":%d}`,
		userID, quizID, attemptNum, score, xpEarned)
	h.db.Insert("user_quiz_attempts", []byte(insertData), true)
}

// ============================================================
// 10. STREAK BONUS XP
// ============================================================

func (h *Handler) getStreakBonusXP(userID string) (int, string) {
	body, err := h.db.Select("user_profiles",
		fmt.Sprintf("id=eq.%s&select=current_streak", userID), true)
	if err != nil {
		return 0, ""
	}
	var profiles []map[string]interface{}
	json.Unmarshal(body, &profiles)
	if len(profiles) == 0 {
		return 0, ""
	}

	streak, _ := profiles[0]["current_streak"].(float64)
	bonus := int(streak) * 5
	msg := ""
	if int(streak) >= 3 {
		msg = fmt.Sprintf("🔥 Série de %d jours ! +%d XP bonus", int(streak), bonus)
	}
	return bonus, msg
}

// ============================================================
// 11. ADMIN TICKETS (from admin_conversations)
// ============================================================

func (h *Handler) AdminTicketsFromDB(c *fiber.Ctx) error {
	body, err := h.db.Select("admin_conversations",
		"order=created_at.desc&limit=50", true)
	if err != nil {
		return renderPage(c, "Tickets", `<p class="text-muted">Erreur de chargement</p>`)
	}
	var conversations []map[string]interface{}
	json.Unmarshal(body, &conversations)

	items := ""
	for _, conv := range conversations {
		id, _ := conv["id"].(string)
		subject, _ := conv["subject"].(string)
		status, _ := conv["status"].(string)
		createdAt, _ := conv["created_at"].(string)

		bc := "badge-outline"
		st := status
		if status == "open" {
			bc = "badge-primary"
			st = "Ouvert"
		} else if status == "assigned" {
			bc = "badge-success"
			st = "Assigné"
		} else if status == "closed" {
			st = "Fermé"
		}

		userID, _ := conv["user_id"].(string)
		username := "User"
		pBody, err := h.db.Select("user_profiles",
			fmt.Sprintf("id=eq.%s&select=username", userID), false)
		if err == nil {
			var profiles []map[string]interface{}
			json.Unmarshal(pBody, &profiles)
			if len(profiles) > 0 {
				if un, ok := profiles[0]["username"].(string); ok {
					username = un
				}
			}
		}

		items += fmt.Sprintf(`
<div class="fr-card">
    <a href="/admin/tickets/%s" class="fr-card-main">
        <div class="fr-avatar">%s</div>
        <div class="fr-info">
            <div class="fr-name">%s</div>
            <div class="fr-meta text-sm text-muted">%s</div>
        </div>
    </a>
    <div class="fr-actions"><span class="%s">%s</span></div>
</div>`, id, string([]byte{username[0]}), subject, createdAt[:10], bc, st)
	}

	if items == "" {
		items = `<p class="text-muted">Aucun ticket</p>`
	}

	return renderPage(c, "Tickets", fmt.Sprintf(`
<div style="max-width:600px;margin:0 auto">
    <h1 class="page-title">🎫 Tickets Support</h1>
    %s
</div>
	`, items))
}

// ============================================================
// FORUM - COMMUNITY CHAT + IDEAS
// ============================================================

func (h *Handler) ForumPage(c *fiber.Ctx) error {
	user := c.Locals("user").(*UserProfile)

	nickname := ""
	if user.Nickname != nil && *user.Nickname != "" {
		nickname = *user.Nickname
	}
	displayName := user.Username
	if nickname != "" {
		displayName = nickname
	}

	forumHTML := `
<div class="forum-page">
    <div class="forum-header">
        <div class="forum-header-left">
            <span class="forum-header-icon">💬</span>
            <div>
                <h1>Discussion Générale</h1>
                <span class="forum-header-subtitle">Communauté Otaku Quiz Africa</span>
            </div>
        </div>
        <span class="forum-online-count" id="forum-online-count">● en ligne</span>
    </div>
    <div class="forum-chat">
        <div class="forum-messages" id="forum-messages">
            <div style="color:#94a3b8;text-align:center;padding:40px">Chargement...</div>
        </div>
        <div class="forum-input">
            <div class="forum-input-wrapper">
                <input type="text" id="forum-input" placeholder="Écrire un message..." autocomplete="off">
                <button id="forum-send-btn" onclick="sendForumMsg()">Envoyer</button>
            </div>
        </div>
    </div>
</div>

<script>
function statsBadge(u){
    if(!u)return'';
    var p=parseInt(u.challenges_played)||0,w=parseInt(u.challenges_won)||0;
    if(p===0)return'';
    return'<span style="display:inline-block;padding:1px 6px;border-radius:4px;background:#6366f122;color:#a78bfa;font-size:.7rem;font-weight:600;margin-left:4px;vertical-align:middle">🏅 '+w+'/'+p+'</span>';
}
var currentChannelId=null;
var forumUserId='__USER_ID__';
var forumDisplayName='__DISPLAY_NAME__';

function initForum(){
    fetch('/api/forum/channels').then(function(r){return r.json()}).then(function(d){
        if(!d.channels)return;
        for(var i=0;i<d.channels.length;i++){
            if(d.channels[i].name.indexOf('💬')!==-1||d.channels[i].name.indexOf('Discussion')!==-1){
                currentChannelId=d.channels[i].id;
                break;
            }
        }
        if(currentChannelId)loadForumMessages();
    });
}

function loadForumMessages(){
    if(!currentChannelId)return;
    var el=document.getElementById('forum-messages');
    fetch('/api/forum/messages?channel_id='+currentChannelId).then(function(r){return r.json()}).then(function(d){
        if(!d.messages||d.messages.length===0){
            el.innerHTML='<div class="forum-empty"><div class="forum-empty-icon">💬</div><div class="forum-empty-text">Aucun message pour l\'instant</div><div class="forum-empty-hint">Soyez le premier à écrire !</div></div>';
            return;
        }
        var h='';
        var lastDate='';
        for(var i=0;i<d.messages.length;i++){
            var m=d.messages[i];
            var msgDate=new Date(m.created_at);
            var dateKey=msgDate.toLocaleDateString('fr-FR');
            if(dateKey!==lastDate){
                var dateLabel=getDateLabel(msgDate);
                h+='<div class="forum-date-sep"><span>'+dateLabel+'</span></div>';
                lastDate=dateKey;
            }
            h+=renderMessage(m,i===d.messages.length-1);
        }
        el.innerHTML=h;
        el.scrollTop=el.scrollHeight;
    });
}

function renderMessage(m,isLast){
    var name='?',avatar='';
    if(m.user){name=m.user.nickname||m.user.username||'?';avatar=m.user.avatar_url||'';}
    var isMine=m.user_id===forumUserId;
    var initial=(name[0]||'?').toUpperCase();
    var time=new Date(m.created_at).toLocaleTimeString('fr-FR',{hour:'2-digit',minute:'2-digit'});
    var edited=m.updated_at?' <span class="forum-msg-edited">(modifié)</span>':'';
    var avatarHtml=avatar?'<img src="'+avatar+'" class="forum-msg-avatar-img">':'<div class="forum-msg-avatar">'+initial+'</div>';

    var h='<div class="forum-msg'+(isMine?' mine':'')+'" id="msg-'+m.id+'">';
    if(!isMine)h+=avatarHtml;
    h+='<div class="forum-msg-body'+(isMine?' mine':'')+'">';
    h+='<div class="forum-msg-top">';
    if(!isMine)h+='<span class="forum-msg-name">'+name+statsBadge(m.user)+'</span>';
    if(isMine)h+='<span class="forum-msg-name mine">Moi</span>';
    h+='<span class="forum-msg-time">'+time+'</span>'+edited;
    h+='</div>';
    if(m.isEditing){
        h+='<div class="forum-msg-edit-form"><input type="text" id="edit-input-'+m.id+'" class="forum-edit-input" value="'+escapeJs(m.content)+'"><button class="forum-edit-save" onclick="saveEdit(\''+m.id+'\')">💾</button><button class="forum-edit-cancel" onclick="cancelEdit(\''+m.id+'\')">❌</button></div>';
    }else{
        h+='<div class="forum-msg-text">'+escapeHtml(m.content)+'</div>';
        h+='<div class="forum-msg-actions">';
        if(isMine){h+='<button class="forum-action-btn" onclick="editMessage(\''+m.id+'\')" title="Modifier">✏️</button><button class="forum-action-btn" onclick="deleteMessage(\''+m.id+'\')" title="Supprimer">🗑️</button>';}
        h+='<button class="forum-action-btn" onclick="reportMessage(\''+m.id+'\')" title="Signaler">🚩</button>';
        h+='</div>';
    }
    h+='</div>';
    if(isMine)h+=avatarHtml;
    h+='</div>';
    return h;
}

function sendForumMsg(){
    var input=document.getElementById('forum-input');
    var content=input.value.trim();
    if(!content||!currentChannelId)return;
    input.value='';
    document.getElementById('forum-send-btn').disabled=true;
    fetch('/api/forum/messages/send',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({channel_id:currentChannelId,content:content})}).then(function(r){return r.json()}).then(function(d){
        document.getElementById('forum-send-btn').disabled=false;
        if(d.success){loadForumMessages();}else{console.error('Send error:',d);input.value=content;}
    }).catch(function(e){console.error('Send error:',e);input.value=content;document.getElementById('forum-send-btn').disabled=false;});
}

var editingMsgId=null;
function editMessage(id){
    if(editingMsgId)cancelEdit(editingMsgId);
    editingMsgId=id;
    var card=document.getElementById('msg-'+id);
    if(!card)return;
    var textEl=card.querySelector('.forum-msg-text');
    var actionsEl=card.querySelector('.forum-msg-actions');
    if(!textEl)return;
    var text=textEl.textContent;
    textEl.style.display='none';
    if(actionsEl)actionsEl.style.display='none';
    var editHtml='<div class="forum-msg-edit-form"><input type="text" id="edit-input-'+id+'" class="forum-edit-input" value="'+escapeJs(text)+'"><button class="forum-edit-save" onclick="saveEdit(\''+id+'\')">💾</button><button class="forum-edit-cancel" onclick="cancelEdit(\''+id+'\')">❌</button></div>';
    var body=card.querySelector('.forum-msg-body');
    var top=body.querySelector('.forum-msg-top');
    top.insertAdjacentHTML('afterend',editHtml);
}

function saveEdit(id){
    var input=document.getElementById('edit-input-'+id);
    if(!input)return;
    var content=input.value.trim();
    if(!content)return;
    fetch('/api/forum/messages/update',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({message_id:id,content:content})}).then(function(r){return r.json()}).then(function(d){
        if(d.success){editingMsgId=null;loadForumMessages();}else{alert(d.error||'Erreur');}
    });
}

function cancelEdit(id){
    editingMsgId=null;
    loadForumMessages();
}

function deleteMessage(id){
    if(!confirm('Supprimer ce message ?'))return;
    fetch('/api/forum/messages/delete',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({message_id:id})}).then(function(r){return r.json()}).then(function(d){
        if(d.success)loadForumMessages();
    });
}

function reportMessage(id){
    var reason=prompt('Raison du signalement :');
    if(!reason||!reason.trim())return;
    fetch('/api/forum/messages/report',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({message_id:id,reason:reason.trim()})}).then(function(r){return r.json()}).then(function(d){
        if(d.success){alert('Signalement envoyé. Merci !');}else{alert(d.error||'Erreur');}
    });
}

function getDateLabel(d){
    var today=new Date();
    var yesterday=new Date(today);yesterday.setDate(yesterday.getDate()-1);
    if(d.toDateString()===today.toDateString())return "Aujourd'hui";
    if(d.toDateString()===yesterday.toDateString())return "Hier";
    return d.toLocaleDateString('fr-FR',{weekday:'long',day:'numeric',month:'long',year:'numeric'});
}

function escapeHtml(t){var d=document.createElement('div');d.textContent=t;return d.innerHTML;}
function escapeJs(s){return s.replace(/'/g,"\\'").replace(/"/g,"&quot;").replace(/</g,"&lt;").replace(/>/g,"&gt;");}

document.getElementById('forum-input').addEventListener('keydown',function(e){if(e.key==='Enter')sendForumMsg();});
initForum();
setInterval(function(){if(currentChannelId)loadForumMessages()},5000);
</script>`

	forumHTML = strings.ReplaceAll(forumHTML, "__USER_ID__", user.ID)
	forumHTML = strings.ReplaceAll(forumHTML, "__DISPLAY_NAME__", displayName)

	content := `
<style>
.forum-page{max-width:800px;margin:0 auto;display:flex;flex-direction:column;height:calc(100vh - 180px);min-height:500px}
.forum-header{display:flex;justify-content:space-between;align-items:center;margin-bottom:16px;flex-shrink:0}
.forum-header-left{display:flex;align-items:center;gap:12px}
.forum-header-icon{font-size:1.8rem}
.forum-header h1{margin:0;font-size:1.3rem;color:white}
.forum-header-subtitle{font-size:.8rem;color:#64748b}
.forum-online-count{font-size:.8rem;color:#22c55e;background:rgba(34,197,94,.1);padding:4px 12px;border-radius:20px}
.forum-chat{flex:1;background:#0f172a;border:1px solid #1e293b;border-radius:16px;display:flex;flex-direction:column;overflow:hidden}
.forum-messages{flex:1;overflow-y:auto;padding:16px 20px;display:flex;flex-direction:column;gap:2px;scroll-behavior:smooth}
.forum-messages::-webkit-scrollbar{width:4px}
.forum-messages::-webkit-scrollbar-track{background:transparent}
.forum-messages::-webkit-scrollbar-thumb{background:#334155;border-radius:4px}
.forum-empty{text-align:center;padding:60px 20px;color:#64748b}
.forum-empty-icon{font-size:3rem;margin-bottom:12px}
.forum-empty-text{font-size:1.1rem;color:#94a3b8;margin-bottom:4px}
.forum-empty-hint{font-size:.85rem}
.forum-date-sep{text-align:center;margin:16px 0 8px}
.forum-date-sep span{display:inline-block;background:#1e293b;color:#94a3b8;font-size:.75rem;padding:4px 14px;border-radius:12px;font-weight:500}
.forum-msg{display:flex;gap:8px;align-items:flex-end;margin-bottom:4px;max-width:85%;animation:fadeIn .2s ease}
@keyframes fadeIn{from{opacity:0;transform:translateY(4px)}to{opacity:1;transform:translateY(0)}}
.forum-msg.mine{margin-left:auto;flex-direction:row-reverse}
.forum-msg-avatar,.forum-msg-avatar-img{width:34px;height:34px;border-radius:50%;flex-shrink:0}
.forum-msg-avatar{background:#6366f1;display:flex;align-items:center;justify-content:center;font-weight:700;color:white;font-size:.75rem}
.forum-msg-avatar-img{object-fit:cover}
.forum-msg-body{background:#1e293b;padding:8px 14px;border-radius:16px 16px 16px 4px;position:relative;min-width:100px}
.forum-msg-body.mine{background:#6366f1;border-radius:16px 16px 4px 16px}
.forum-msg-top{display:flex;align-items:center;gap:6px;margin-bottom:2px;flex-wrap:wrap}
.forum-msg-name{font-weight:600;font-size:.8rem;color:#a78bfa}
.forum-msg-name.mine{color:#c7d2fe}
.forum-msg-time{font-size:.65rem;color:#64748b}
.forum-msg-body.mine .forum-msg-time{color:#a5b4fc}
.forum-msg-edited{font-size:.65rem;color:#64748b;font-style:italic}
.forum-msg-text{color:#e2e8f0;font-size:.9rem;line-height:1.45;word-wrap:break-word;white-space:pre-wrap}
.forum-msg-body.mine .forum-msg-text{color:white}
.forum-msg-actions{display:flex;gap:2px;margin-top:4px;opacity:0;transition:opacity .15s}
.forum-msg:hover .forum-msg-actions{opacity:1}
.forum-action-btn{background:none;border:none;color:#64748b;cursor:pointer;font-size:.75rem;padding:2px 6px;border-radius:6px;transition:all .15s}
.forum-action-btn:hover{background:#334155;color:#e2e8f0}
.forum-msg-edit-form{display:flex;gap:6px;margin-top:4px}
.forum-edit-input{flex:1;padding:6px 10px;border-radius:8px;border:1px solid #6366f1;background:#0f172a;color:white;font-size:.85rem;outline:none}
.forum-edit-save,.forum-edit-cancel{background:none;border:none;cursor:pointer;font-size:.9rem;padding:2px 6px;border-radius:6px}
.forum-edit-save:hover{background:rgba(34,197,94,.2)}
.forum-edit-cancel:hover{background:rgba(239,68,68,.2)}
.forum-input{flex-shrink:0;padding:12px 16px;border-top:1px solid #1e293b}
.forum-input-wrapper{display:flex;gap:8px;align-items:center}
.forum-input-wrapper input{flex:1;padding:12px 16px;border-radius:24px;border:1px solid #334155;background:#1e293b;color:white;font-size:.9rem;outline:none;transition:border-color .2s}
.forum-input-wrapper input:focus{border-color:#6366f1}
.forum-input-wrapper button{padding:10px 24px;border-radius:24px;background:#6366f1;color:white;border:none;font-weight:600;cursor:pointer;font-size:.9rem;transition:background .2s;white-space:nowrap}
.forum-input-wrapper button:hover{background:#4f46e5}
.forum-input-wrapper button:disabled{opacity:.5;cursor:not-allowed}
</style>` + forumHTML

	return renderPage(c, "Forum", content)
}

func (h *Handler) APIGetForumChannels(c *fiber.Ctx) error {
	body, err := h.db.Select("forum_channels", "order=name.asc", true)
	if err != nil {
		return c.JSON(fiber.Map{"channels": []interface{}{}})
	}
	var channels []map[string]interface{}
	json.Unmarshal(body, &channels)
	return c.JSON(fiber.Map{"channels": channels})
}

func (h *Handler) APIGetForumMessages(c *fiber.Ctx) error {
	channelID := c.Query("channel_id")
	if channelID == "" {
		return c.JSON(fiber.Map{"messages": []interface{}{}})
	}

	body, err := h.db.Select("forum_messages",
		fmt.Sprintf("channel_id=eq.%s&order=created_at.asc&limit=200", channelID), true)
	if err != nil {
		return c.JSON(fiber.Map{"messages": []interface{}{}, "error": err.Error()})
	}
	var messages []map[string]interface{}
	json.Unmarshal(body, &messages)
	log.Printf("[APIGetForumMessages] channel=%s found %d messages", channelID, len(messages))

	userIDs := map[string]bool{}
	for _, m := range messages {
		if uid, ok := m["user_id"].(string); ok {
			userIDs[uid] = true
		}
	}
	ids := []string{}
	for id := range userIDs {
		ids = append(ids, id)
	}
	profilesMap := map[string]map[string]interface{}{}
	if len(ids) > 0 {
		profiles, _ := h.db.GetProfiles(ids)
		for _, p := range profiles {
			profilesMap[database.DBValue(p["id"])] = p
		}
	}

	type enrichedMsg struct {
		ID        string                 `json:"id"`
		ChannelID string                 `json:"channel_id"`
		UserID    string                 `json:"user_id"`
		Content   string                 `json:"content"`
		CreatedAt string                 `json:"created_at"`
		UpdatedAt string                 `json:"updated_at"`
		User      map[string]interface{} `json:"user"`
	}
	enriched := []enrichedMsg{}
	for _, m := range messages {
		uid := database.DBValue(m["user_id"])
		em := enrichedMsg{
			ID:        database.DBValue(m["id"]),
			ChannelID: database.DBValue(m["channel_id"]),
			UserID:    uid,
			Content:   database.DBValue(m["content"]),
			CreatedAt: database.DBValue(m["created_at"]),
			UpdatedAt: database.DBValue(m["updated_at"]),
			User:      profilesMap[uid],
		}
		enriched = append(enriched, em)
	}

	return c.JSON(fiber.Map{"messages": enriched})
}

func (h *Handler) APISendForumMessage(c *fiber.Ctx) error {
	type Request struct {
		ChannelID string `json:"channel_id"`
		Content   string `json:"content"`
	}
	var req Request
	c.BodyParser(&req)

	sess, _ := h.store.Get(c)
	userID := sess.Get("user_id").(string)

	if req.ChannelID == "" || req.Content == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Données invalides"})
	}

	msgData, _ := json.Marshal(map[string]interface{}{
		"channel_id": req.ChannelID,
		"user_id":    userID,
		"content":    req.Content,
	})
	_, err := h.db.Insert("forum_messages", msgData, true)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true})
}

func (h *Handler) APIUpdateForumMessage(c *fiber.Ctx) error {
	type Request struct {
		MessageID string `json:"message_id"`
		Content   string `json:"content"`
	}
	var req Request
	c.BodyParser(&req)

	sess, _ := h.store.Get(c)
	userID := sess.Get("user_id").(string)

	if req.MessageID == "" || req.Content == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Données invalides"})
	}

	msgBody, _ := h.db.Select("forum_messages",
		fmt.Sprintf("id=eq.%s&user_id=eq.%s&select=id", req.MessageID, userID), true)
	var msgs []map[string]interface{}
	json.Unmarshal(msgBody, &msgs)
	if len(msgs) == 0 {
		return c.Status(403).JSON(fiber.Map{"error": "Non autorisé"})
	}

	data, _ := json.Marshal(map[string]interface{}{
		"content":    req.Content,
		"updated_at": time.Now().UTC().Format(time.RFC3339),
	})
	_, err := h.db.Update("forum_messages",
		fmt.Sprintf("id=eq.%s", req.MessageID), data, true)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true})
}

func (h *Handler) APIDeleteForumMessage(c *fiber.Ctx) error {
	type Request struct {
		MessageID string `json:"message_id"`
	}
	var req Request
	c.BodyParser(&req)

	sess, _ := h.store.Get(c)
	userID := sess.Get("user_id").(string)

	if req.MessageID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "ID requis"})
	}

	msgBody, _ := h.db.Select("forum_messages",
		fmt.Sprintf("id=eq.%s&user_id=eq.%s&select=id", req.MessageID, userID), true)
	var msgs []map[string]interface{}
	json.Unmarshal(msgBody, &msgs)
	if len(msgs) == 0 {
		return c.Status(403).JSON(fiber.Map{"error": "Non autorisé"})
	}

	h.db.Delete("forum_messages",
		fmt.Sprintf("id=eq.%s", req.MessageID), true)
	return c.JSON(fiber.Map{"success": true})
}

func (h *Handler) APIReportForumMessage(c *fiber.Ctx) error {
	type Request struct {
		MessageID string `json:"message_id"`
		Reason    string `json:"reason"`
	}
	var req Request
	c.BodyParser(&req)

	sess, _ := h.store.Get(c)
	userID := sess.Get("user_id").(string)

	if req.MessageID == "" || req.Reason == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Données invalides"})
	}

	data, _ := json.Marshal(map[string]interface{}{
		"message_id":  req.MessageID,
		"reporter_id": userID,
		"reason":      req.Reason,
	})
	_, err := h.db.Insert("forum_message_reports", data, true)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true})
}

// ============================================================
// FORUM SUGGESTIONS — BOÎTE À IDÉES
// ============================================================

func currentWeekLabel() string {
	now := time.Now()
	year, week := now.ISOWeek()
	return fmt.Sprintf("%d-W%02d", year, week)
}

func (h *Handler) APIGetForumSuggestions(c *fiber.Ctx) error {
	channelID := c.Query("channel_id")
	if channelID == "" {
		return c.JSON(fiber.Map{"suggestions": []interface{}{}})
	}

	body, err := h.db.Select("forum_suggestions",
		fmt.Sprintf("channel_id=eq.%s&order=created_at.desc", channelID), true)
	if err != nil {
		return c.JSON(fiber.Map{"suggestions": []interface{}{}, "error": err.Error()})
	}
	var suggestions []map[string]interface{}
	json.Unmarshal(body, &suggestions)

	type enrichedSuggestion struct {
		ID        string                 `json:"id"`
		ChannelID string                 `json:"channel_id"`
		UserID    string                 `json:"user_id"`
		Title     string                 `json:"title"`
		Content   string                 `json:"content"`
		WeekLabel string                 `json:"week_label"`
		CreatedAt string                 `json:"created_at"`
		User      map[string]interface{} `json:"user"`
		UpVotes   int                    `json:"up_votes"`
		DownVotes int                    `json:"down_votes"`
		MyVote    string                 `json:"my_vote"`
	}

	sess, _ := h.store.Get(c)
	currentUserID := ""
	if uid := sess.Get("user_id"); uid != nil {
		currentUserID = uid.(string)
	}

	userIDs := map[string]bool{}
	for _, s := range suggestions {
		if uid, ok := s["user_id"].(string); ok {
			userIDs[uid] = true
		}
	}
	ids := []string{}
	for id := range userIDs {
		ids = append(ids, id)
	}
	profilesMap := map[string]map[string]interface{}{}
	if len(ids) > 0 {
		profiles, _ := h.db.GetProfiles(ids)
		for _, p := range profiles {
			profilesMap[database.DBValue(p["id"])] = p
		}
	}

	enriched := []enrichedSuggestion{}
	for _, s := range suggestions {
		sID := database.DBValue(s["id"])
		uid := database.DBValue(s["user_id"])

		upBody, _ := h.db.Select("forum_suggestion_votes",
			fmt.Sprintf("suggestion_id=eq.%s&vote_type=eq.up&select=id", sID), true)
		var ups []map[string]interface{}
		json.Unmarshal(upBody, &ups)

		downBody, _ := h.db.Select("forum_suggestion_votes",
			fmt.Sprintf("suggestion_id=eq.%s&vote_type=eq.down&select=id", sID), true)
		var downs []map[string]interface{}
		json.Unmarshal(downBody, &downs)

		myVote := ""
		if currentUserID != "" {
			voteBody, _ := h.db.Select("forum_suggestion_votes",
				fmt.Sprintf("suggestion_id=eq.%s&user_id=eq.%s&select=vote_type", sID, currentUserID), true)
			var myVotes []map[string]interface{}
			json.Unmarshal(voteBody, &myVotes)
			if len(myVotes) > 0 {
				myVote, _ = myVotes[0]["vote_type"].(string)
			}
		}

		enriched = append(enriched, enrichedSuggestion{
			ID:        sID,
			ChannelID: database.DBValue(s["channel_id"]),
			UserID:    uid,
			Title:     database.DBValue(s["title"]),
			Content:   database.DBValue(s["content"]),
			WeekLabel: database.DBValue(s["week_label"]),
			CreatedAt: database.DBValue(s["created_at"]),
			User:      profilesMap[uid],
			UpVotes:   len(ups),
			DownVotes: len(downs),
			MyVote:    myVote,
		})
	}

	return c.JSON(fiber.Map{"suggestions": enriched})
}

func (h *Handler) APICreateForumSuggestion(c *fiber.Ctx) error {
	type Request struct {
		ChannelID string `json:"channel_id"`
		Title     string `json:"title"`
		Content   string `json:"content"`
	}
	var req Request
	c.BodyParser(&req)

	sess, _ := h.store.Get(c)
	userID := sess.Get("user_id").(string)

	log.Printf("[APICreateForumSuggestion] user=%s channel_id=%q title=%q content=%q", userID, req.ChannelID, req.Title, req.Content)

	if req.ChannelID == "" || req.Title == "" || req.Content == "" {
		log.Printf("[APICreateForumSuggestion] validation failed: channel_id=%q title=%q content=%q", req.ChannelID, req.Title, req.Content)
		return c.Status(400).JSON(fiber.Map{"error": "Titre et contenu requis"})
	}

	data, _ := json.Marshal(map[string]interface{}{
		"channel_id": req.ChannelID,
		"user_id":    userID,
		"title":      req.Title,
		"content":    req.Content,
		"week_label": currentWeekLabel(),
	})
	log.Printf("[APICreateForumSuggestion] inserting: %s", string(data))
	_, err := h.db.Insert("forum_suggestions", data, true)
	if err != nil {
		log.Printf("[APICreateForumSuggestion] insert error: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	log.Printf("[APICreateForumSuggestion] insert success")
	return c.JSON(fiber.Map{"success": true})
}

func (h *Handler) APISuggestionVote(c *fiber.Ctx) error {
	type Request struct {
		SuggestionID string `json:"suggestion_id"`
		VoteType     string `json:"vote_type"`
	}
	var req Request
	c.BodyParser(&req)

	sess, _ := h.store.Get(c)
	userID := sess.Get("user_id").(string)

	if req.SuggestionID == "" || (req.VoteType != "up" && req.VoteType != "down") {
		return c.Status(400).JSON(fiber.Map{"error": "Données invalides"})
	}

	checkBody, _ := h.db.Select("forum_suggestion_votes",
		fmt.Sprintf("suggestion_id=eq.%s&user_id=eq.%s&select=id,vote_type", req.SuggestionID, userID), true)
	var existing []map[string]interface{}
	json.Unmarshal(checkBody, &existing)

	if len(existing) > 0 {
		oldVote, _ := existing[0]["vote_type"].(string)
		if oldVote == req.VoteType {
			h.db.Delete("forum_suggestion_votes",
				fmt.Sprintf("suggestion_id=eq.%s&user_id=eq.%s", req.SuggestionID, userID), true)
		} else {
			h.db.Update("forum_suggestion_votes",
				fmt.Sprintf("suggestion_id=eq.%s&user_id=eq.%s", req.SuggestionID, userID),
				[]byte(fmt.Sprintf(`{"vote_type":"%s"}`, req.VoteType)), true)
		}
	} else {
		voteData, _ := json.Marshal(map[string]interface{}{
			"suggestion_id": req.SuggestionID,
			"user_id":       userID,
			"vote_type":     req.VoteType,
		})
		h.db.Insert("forum_suggestion_votes", voteData, true)
	}

	upBody, _ := h.db.Select("forum_suggestion_votes",
		fmt.Sprintf("suggestion_id=eq.%s&vote_type=eq.up&select=id", req.SuggestionID), true)
	var ups []map[string]interface{}
	json.Unmarshal(upBody, &ups)

	downBody, _ := h.db.Select("forum_suggestion_votes",
		fmt.Sprintf("suggestion_id=eq.%s&vote_type=eq.down&select=id", req.SuggestionID), true)
	var downs []map[string]interface{}
	json.Unmarshal(downBody, &downs)

	return c.JSON(fiber.Map{
		"success":    true,
		"up_votes":   len(ups),
		"down_votes": len(downs),
	})
}

func (h *Handler) APIDeleteForumSuggestion(c *fiber.Ctx) error {
	sess, _ := h.store.Get(c)
	userID := sess.Get("user_id").(string)

	userBody, _ := h.db.Select("user_profiles",
		fmt.Sprintf("id=eq.%s&select=is_admin", userID), true)
	var users []map[string]interface{}
	json.Unmarshal(userBody, &users)
	isAdmin := false
	if len(users) > 0 {
		isAdmin = database.DBBool(users[0]["is_admin"])
	}

	type Request struct {
		SuggestionID string `json:"suggestion_id"`
	}
	var req Request
	c.BodyParser(&req)

	if req.SuggestionID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "ID requis"})
	}

	if !isAdmin {
		sugBody, _ := h.db.Select("forum_suggestions",
			fmt.Sprintf("id=eq.%s&user_id=eq.%s&select=id", req.SuggestionID, userID), true)
		var sugs []map[string]interface{}
		json.Unmarshal(sugBody, &sugs)
		if len(sugs) == 0 {
			return c.Status(403).JSON(fiber.Map{"error": "Non autorisé"})
		}
	}

	h.db.Delete("forum_suggestion_votes",
		fmt.Sprintf("suggestion_id=eq.%s", req.SuggestionID), true)
	h.db.Delete("forum_suggestions",
		fmt.Sprintf("id=eq.%s", req.SuggestionID), true)

	return c.JSON(fiber.Map{"success": true})
}

func (h *Handler) APISuggestionUpdate(c *fiber.Ctx) error {
	sess, _ := h.store.Get(c)
	userID := sess.Get("user_id").(string)

	type Request struct {
		SuggestionID string `json:"suggestion_id"`
		Title        string `json:"title"`
		Content      string `json:"content"`
	}
	var req Request
	c.BodyParser(&req)

	if req.SuggestionID == "" || req.Title == "" || req.Content == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Données invalides"})
	}

	sugBody, _ := h.db.Select("forum_suggestions",
		fmt.Sprintf("id=eq.%s&user_id=eq.%s&select=id", req.SuggestionID, userID), true)
	var sugs []map[string]interface{}
	json.Unmarshal(sugBody, &sugs)
	if len(sugs) == 0 {
		return c.Status(403).JSON(fiber.Map{"error": "Non autorisé"})
	}

	data, _ := json.Marshal(map[string]interface{}{
		"title":   req.Title,
		"content": req.Content,
	})
	_, err := h.db.Update("forum_suggestions",
		fmt.Sprintf("id=eq.%s", req.SuggestionID), data, true)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true})
}

func (h *Handler) IdeasBoxPage(c *fiber.Ctx) error {
	user := c.Locals("user").(*UserProfile)

	isAdminStr := "false"
	if user.IsAdmin {
		isAdminStr = "true"
	}

	ideasChannelID := ""
	chBody, _ := h.db.Select("forum_channels",
		"select=id,name", true)
	log.Printf("[IdeasBoxPage] forum_channels raw body: %s", string(chBody))
	var chRows []map[string]interface{}
	json.Unmarshal(chBody, &chRows)
	log.Printf("[IdeasBoxPage] parsed %d channels", len(chRows))
	for _, ch := range chRows {
		chName := database.DBValue(ch["name"])
		chID := database.DBValue(ch["id"])
		log.Printf("[IdeasBoxPage] channel: id=%s name=%s", chID, chName)
		if strings.Contains(chName, "Boîte à Idées") {
			ideasChannelID = chID
			log.Printf("[IdeasBoxPage] MATCHED! channel_id=%s", ideasChannelID)
			break
		}
	}
	if ideasChannelID == "" {
		log.Printf("[IdeasBoxPage] WARNING: no Boîte à Idées channel found!")
	}

	pageHTML := `
<div class="ideasbox-page">
    <div class="ideasbox-header">
        <div class="ideasbox-header-left">
            <h1>💡 Boîte à Idées</h1>
            <span class="ideasbox-subtitle">Propose et vote pour les améliorations de la plateforme</span>
        </div>
        <button class="btn-primary" onclick="showIdeaForm()">+ Idée</button>
    </div>

    <div id="idea-form" class="idea-form-card" style="display:none">
        <input type="text" id="idea-title-input" placeholder="Titre de ton idée..." maxlength="200">
        <textarea id="idea-content-input" placeholder="Décris ton idée en détail..." maxlength="2000" rows="3"></textarea>
        <div class="idea-form-actions">
            <button class="btn-outline" onclick="hideIdeaForm()">Annuler</button>
            <button class="btn-primary" onclick="submitIdea()">Publier</button>
        </div>
    </div>

    <div id="ideas-container"></div>
</div>

<script>
function statsBadge(u){
    if(!u)return'';
    var p=parseInt(u.challenges_played)||0,w=parseInt(u.challenges_won)||0;
    if(p===0)return'';
    return'<span style="display:inline-block;padding:1px 6px;border-radius:4px;background:#6366f122;color:#a78bfa;font-size:.7rem;font-weight:600;margin-left:4px;vertical-align:middle">🏅 '+w+'/'+p+'</span>';
}
var ideasUserId='__USER_ID__';
var ideasIsAdmin='__IS_ADMIN__';

function showIdeaForm(){document.getElementById('idea-form').style.display='block';document.getElementById('idea-title-input').focus();}
function hideIdeaForm(){document.getElementById('idea-form').style.display='none';document.getElementById('idea-title-input').value='';document.getElementById('idea-content-input').value='';}

function submitIdea(){
    var t=document.getElementById('idea-title-input').value.trim();
    var c=document.getElementById('idea-content-input').value.trim();
    if(!t||!c){alert('Remplis le titre et le contenu');return;}
    fetch('/api/forum/suggestions/create',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({channel_id:'__CHANNEL_ID__',title:t,content:c})}).then(function(r){return r.json()}).then(function(d){
        console.log('[submitIdea] response:',JSON.stringify(d));
        if(d.success){hideIdeaForm();loadIdeas();}else{alert(d.error||'Erreur');}
    }).catch(function(e){console.log('[submitIdea] fetch error:',e);alert('Erreur réseau');});
}

function loadIdeas(){
    var el=document.getElementById('ideas-container');
    el.innerHTML='<div style="text-align:center;color:#64748b;padding:40px">Chargement...</div>';
    fetch('/api/forum/suggestions/all').then(function(r){return r.json()}).then(function(d){
        console.log('[loadIdeas] response:',JSON.stringify(d));
        if(!d.suggestions||d.suggestions.length===0){el.innerHTML='<div style="text-align:center;color:#64748b;padding:60px 20px"><div style="font-size:2rem;margin-bottom:12px">💡</div>Aucune idée pour l\'instant.<br>Soyez le premier à proposer !</div>';return;}
        var weeks={};
        for(var i=0;i<d.suggestions.length;i++){
            var s=d.suggestions[i];
            var w=s.week_label||'sans-semaine';
            if(!weeks[w])weeks[w]=[];
            weeks[w].push(s);
        }
        var weekKeys=Object.keys(weeks).sort().reverse();
        var h='';
        for(var wi=0;wi<weekKeys.length;wi++){
            var wk=weekKeys[wi];
            var items=weeks[wk];
            var weekTitle=getWeekTitle(wk);
            h+='<div class="ideasbox-week">';
            h+='<div class="ideasbox-week-header">';
            h+='<span class="ideasbox-week-icon">📅</span>';
            h+='<span>'+weekTitle+'</span>';
            h+='<span class="ideasbox-week-count">'+items.length+'</span>';
            h+='</div>';
            items.sort(function(a,b){return(b.up_votes-b.down_votes)-(a.up_votes-a.down_votes);});
            for(var i=0;i<items.length;i++){
                h+=renderIdeaCard(items[i]);
            }
            h+='</div>';
        }
        el.innerHTML=h;
    });
}

function renderIdeaCard(s){
    var name='?',avatar='';
    if(s.user){name=s.user.nickname||s.user.username||'?';avatar=s.user.avatar_url||'';}
    var initial=(name[0]||'?').toUpperCase();
    var score=s.up_votes-s.down_votes;
    var total=s.up_votes+s.down_votes;
    var upPct=total>0?Math.round(s.up_votes/total*100):0;
    var downPct=total>0?Math.round(s.down_votes/total*100):0;
    var date=new Date(s.created_at).toLocaleDateString('fr-FR',{day:'2-digit',month:'short',year:'numeric'});
    var canDelete=(ideasIsAdmin==='true')||(s.user_id===ideasUserId);
    var canEdit=(s.user_id===ideasUserId);

    var avatarHtml=avatar?'<img src="'+avatar+'" class="idea-avatar-img">':'<div class="idea-avatar">'+initial+'</div>';
    var h='<div class="idea-card" id="idea-'+s.id+'">';
    h+='<div class="idea-card-header">';
    h+='<div class="idea-author-info">';
    h+=avatarHtml;
    h+='<span class="idea-author-name">'+name+statsBadge(s.user)+'</span>';
    h+='<span class="idea-date">'+date+'</span>';
    h+='</div>';
    h+='<div class="idea-card-actions">';
    if(canEdit){h+='<button class="idea-edit-btn" onclick="editIdea(\''+s.id+'\',\''+escapeJs(s.title)+'\',\''+escapeJs(s.content)+'\')">✏️</button>';}
    if(canDelete){h+='<button class="idea-delete-btn" onclick="deleteIdea(\''+s.id+'\')">🗑️</button>';}
    h+='</div>';
    h+='</div>';
    h+='<div class="idea-card-title">'+escapeHtml(s.title)+'</div>';
    h+='<div class="idea-card-content">'+escapeHtml(s.content)+'</div>';
    h+='<div class="idea-card-vote">';
    h+='<div class="vote-row">';
    h+='<button class="vote-option vote-up '+(s.my_vote==='up'?'vote-active-up':'')+'" onclick="voteIdea(\''+s.id+'\',\'up\')">';
    h+='<span class="vote-emoji">👍</span>';
    h+='<span class="vote-bar-wrap"><span class="vote-bar vote-bar-up" style="width:'+upPct+'%"></span></span>';
    h+='<span class="vote-label">'+s.up_votes+' <span class="vote-pct">'+upPct+'%</span></span>';
    h+='</button>';
    h+='</div>';
    h+='<div class="vote-row">';
    h+='<button class="vote-option vote-down '+(s.my_vote==='down'?'vote-active-down':'')+'" onclick="voteIdea(\''+s.id+'\',\'down\')">';
    h+='<span class="vote-emoji">👎</span>';
    h+='<span class="vote-bar-wrap"><span class="vote-bar vote-bar-down" style="width:'+downPct+'%"></span></span>';
    h+='<span class="vote-label">'+s.down_votes+' <span class="vote-pct">'+downPct+'%</span></span>';
    h+='</button>';
    h+='</div>';
    h+='</div></div>';
    return h;
}

function voteIdea(id,type){
    fetch('/api/forum/suggestions/vote',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({suggestion_id:id,vote_type:type})}).then(function(r){return r.json()}).then(function(d){
        if(d.success)loadIdeas();
    });
}

function deleteIdea(id){
    if(!confirm('Supprimer cette idée ?'))return;
    fetch('/api/forum/suggestions/delete',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({suggestion_id:id})}).then(function(r){return r.json()}).then(function(d){
        if(d.success)loadIdeas();
    });
}

var editingId=null;
function editIdea(id,title,content){
    if(editingId)cancelEdit(editingId);
    editingId=id;
    var card=document.getElementById('idea-'+id);
    if(!card)return;
    card.classList.add('editing');
    card.querySelector('.idea-card-title').innerHTML='<input type="text" id="edit-title-'+id+'" class="idea-edit-input" value="'+title+'" maxlength="200">';
    card.querySelector('.idea-card-content').innerHTML='<textarea id="edit-content-'+id+'" class="idea-edit-textarea" maxlength="2000" rows="3">'+content+'</textarea>';
    card.querySelector('.idea-card-actions').innerHTML='<button class="btn-primary" style="padding:6px 14px;font-size:.8rem" onclick="saveIdea(\''+id+'\')">💾</button><button class="btn-outline" style="padding:6px 14px;font-size:.8rem" onclick="cancelEdit(\''+id+'\')">❌</button>';
}
function cancelEdit(id){
    editingId=null;
    var card=document.getElementById('idea-'+id);
    if(!card)return;
    card.classList.remove('editing');
    loadIdeas();
}
function saveIdea(id){
    var title=document.getElementById('edit-title-'+id).value.trim();
    var content=document.getElementById('edit-content-'+id).value.trim();
    if(!title||!content){alert('Le titre et le contenu sont requis');return;}
    fetch('/api/forum/suggestions/update',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({suggestion_id:id,title:title,content:content})}).then(function(r){return r.json()}).then(function(d){
        if(d.success){editingId=null;loadIdeas();}else{alert(d.error||'Erreur');}
    });
}

function getWeekTitle(wl){
    var parts=wl.split('-W');
    if(parts.length<2)return wl;
    var year=parseInt(parts[0]);
    var week=parseInt(parts[1]);
    var jan1=new Date(year,0,1);
    var days=4-jan1.getDay();
    if(days<0)days+=7;
    var d=new Date(year,0,1+days+(week-1)*7);
    var end=new Date(d);
    end.setDate(end.getDate()+6);
    var opts={day:'2-digit',month:'short'};
    return d.toLocaleDateString('fr-FR',opts)+' → '+end.toLocaleDateString('fr-FR',opts)+' '+year;
}

function escapeHtml(t){var d=document.createElement('div');d.textContent=t;return d.innerHTML;}
function escapeJs(s){return s.replace(/'/g,"\\'").replace(/"/g,"&quot;").replace(/</g,"&lt;").replace(/>/g,"&gt;");}

loadIdeas();
</script>`

	pageHTML = strings.ReplaceAll(pageHTML, "__USER_ID__", user.ID)
	pageHTML = strings.ReplaceAll(pageHTML, "__IS_ADMIN__", isAdminStr)
	pageHTML = strings.ReplaceAll(pageHTML, "__CHANNEL_ID__", ideasChannelID)

	content := `
<style>
.ideasbox-page{max-width:800px;margin:0 auto}
.ideasbox-header{display:flex;justify-content:space-between;align-items:center;margin-bottom:24px;flex-wrap:wrap;gap:12px}
.ideasbox-header h1{margin:0;font-size:1.5rem}
.ideasbox-subtitle{color:#64748b;font-size:.85rem}

.idea-form-card{background:#16213e;border:1px solid #2d2d44;border-radius:12px;padding:20px;margin-bottom:24px}
.idea-form-card input,.idea-form-card textarea{width:100%;padding:12px 14px;border-radius:10px;border:1px solid #2d2d44;background:#0f172a;color:white;font-size:.9rem;box-sizing:border-box;margin-bottom:10px;font-family:inherit;resize:vertical}
.idea-form-card input:focus,.idea-form-card textarea:focus{outline:none;border-color:#6366f1}
.idea-form-actions{display:flex;gap:8px;justify-content:flex-end}

.ideasbox-week{margin-bottom:28px}
.ideasbox-week-header{display:flex;align-items:center;gap:8px;padding:8px 0;border-bottom:1px solid #2d2d44;margin-bottom:12px;font-weight:600;font-size:.85rem;color:#94a3b8}
.ideasbox-week-icon{font-size:1rem}
.ideasbox-week-count{margin-left:auto;font-size:.75rem;background:#1e293b;padding:2px 8px;border-radius:4px;color:#64748b}

.idea-card{background:#16213e;border:1px solid #2d2d44;border-radius:12px;padding:18px;margin-bottom:10px;transition:border-color .2s}
.idea-card:hover{border-color:#6366f1}
.idea-card-header{display:flex;justify-content:space-between;align-items:center;margin-bottom:10px}
.idea-author-info{display:flex;align-items:center;gap:8px}
.idea-avatar{width:30px;height:30px;border-radius:50%;background:#6366f1;display:flex;align-items:center;justify-content:center;font-weight:700;color:white;font-size:.75rem;flex-shrink:0}
.idea-avatar-img{width:30px;height:30px;border-radius:50%;object-fit:cover;flex-shrink:0}
.idea-author-name{font-weight:600;font-size:.85rem;color:#e2e8f0}
.idea-date{font-size:.75rem;color:#64748b}
.idea-delete-btn{background:none;border:none;color:#ef4444;cursor:pointer;font-size:.85rem;padding:4px;border-radius:6px;transition:background .2s}
.idea-delete-btn:hover{background:rgba(239,68,68,.15)}
.idea-edit-btn{background:none;border:none;color:#6366f1;cursor:pointer;font-size:.85rem;padding:4px;border-radius:6px;transition:background .2s}
.idea-edit-btn:hover{background:rgba(99,102,241,.15)}
.idea-card-actions{display:flex;gap:4px;align-items:center}
.idea-card.editing .idea-card-title,.idea-card.editing .idea-card-content{padding:0}
.idea-edit-input,.idea-edit-textarea{width:100%;padding:10px 12px;border-radius:8px;border:1px solid #6366f1;background:#0f172a;color:white;font-size:.9rem;box-sizing:border-box;font-family:inherit;resize:vertical;outline:none}
.idea-edit-textarea{min-height:80px}
.idea-card-title{font-weight:700;font-size:1.05rem;margin-bottom:6px;color:white}
.idea-card-content{color:#94a3b8;font-size:.85rem;line-height:1.6;margin-bottom:14px;white-space:pre-wrap}

.idea-card-vote{display:flex;flex-direction:column;gap:6px}
.vote-row{position:relative}
.vote-option{width:100%;display:flex;align-items:center;gap:10px;padding:10px 14px;border-radius:10px;border:2px solid #2d2d44;background:#0f172a;color:white;cursor:pointer;font-size:.85rem;font-weight:500;transition:all .15s;position:relative;overflow:hidden;text-align:left}
.vote-option:hover{border-color:#4a4a6a}
.vote-active-up{border-color:#22c55e;background:rgba(34,197,94,.08)}
.vote-active-down{border-color:#ef4444;background:rgba(239,68,68,.08)}
.vote-emoji{font-size:1.1rem;flex-shrink:0;position:relative;z-index:2}
.vote-bar-wrap{flex:1;height:24px;background:#1a1a3e;border-radius:6px;overflow:hidden;position:relative;z-index:1}
.vote-bar{height:100%;border-radius:6px;transition:width .4s ease}
.vote-bar-up{background:linear-gradient(90deg,#22c55e 0%,#16a34a 100%)}
.vote-bar-down{background:linear-gradient(90deg,#ef4444 0%,#dc2626 100%)}
.vote-label{font-size:.8rem;font-weight:700;position:relative;z-index:2;white-space:nowrap;min-width:80px;text-align:right}
.vote-pct{color:#94a3b8;font-weight:400}
</style>` + pageHTML

	return renderPage(c, "Boîte à Idées", content)
}
