package handlers

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"otaku-quiz-africa/internal/database"
)

// ============================================================
// ANON BOX — CRUD
// ============================================================

func (h *Handler) AnonBoxCreate(c *fiber.Ctx) error {
	user := h.getUserFromSession(c)
	if user == nil {
		return c.Redirect("/login")
	}

	checkBody, _ := h.db.Select("anon_boxes",
		fmt.Sprintf("owner_id=eq.%s&active=eq.true&select=id", user.ID), true)
	var existing []map[string]interface{}
	json.Unmarshal(checkBody, &existing)
	if len(existing) > 0 {
		return c.Redirect("/anon-box")
	}

	token := generateToken(8)

	data, _ := json.Marshal(map[string]interface{}{
		"owner_id": user.ID,
		"token":    token,
		"active":   true,
	})
	h.db.Insert("anon_boxes", data, true)

	return c.Redirect("/anon-box")
}

func (h *Handler) AnonBoxDashboard(c *fiber.Ctx) error {
	user := h.getUserFromSession(c)
	if user == nil {
		return c.Redirect("/login")
	}

	boxBody, _ := h.db.Select("anon_boxes",
		fmt.Sprintf("owner_id=eq.%s&active=eq.true&select=*", user.ID), true)
	var boxes []map[string]interface{}
	json.Unmarshal(boxBody, &boxes)

	content := `<div class="ngl-dash">
<div class="ngl-dash-inner">
<h1 class="ngl-dash-title">💌 Messages Anonymes</h1>`

	if len(boxes) == 0 {
		content += `
<div class="ngl-empty">
<p>Crée ta boîte pour recevoir des messages anonymes de tes amis !</p>
<form method="POST" action="/anon-box/create">
<button class="ngl-btn-primary">Créer ma boîte</button>
</form>
</div>`
	} else {
		box := boxes[0]
		boxID := database.DBValue(box["id"])
		boxToken, _ := box["token"].(string)

		msgsBody, _ := h.db.Select("anon_messages",
			fmt.Sprintf("box_id=eq.%s&select=*&order=sent_at.desc&limit=50", boxID), true)
		log.Printf("[AnonBoxDashboard] msgs body=%s", string(msgsBody))
		var msgs []map[string]interface{}
		json.Unmarshal(msgsBody, &msgs)

		unreadCount := 0
		for _, m := range msgs {
			if !database.DBBool(m["is_read"]) {
				unreadCount++
			}
		}

		content += `
<div class="ngl-share">
<div class="ngl-share-label">Ton lien NGL</div>
<div class="ngl-share-row">
<span class="ngl-prefix">/anon/</span>
<input type="text" id="anon-token" value="` + boxToken + `" oninput="onAnonTokenChange()">
<button class="ngl-btn-copy" onclick="copyAnonLink()" title="Copier">📋</button>
</div>
<div class="ngl-share-actions">
<button class="ngl-btn-ghost" id="anon-save-btn" style="display:none" onclick="saveAnonToken()">💾 Enregistrer le token</button>
<button class="ngl-btn-ghost" onclick="resetAnonLink()">🔄 Nouveau lien</button>
</div>
</div>

<div class="ngl-inbox">
<div class="ngl-inbox-header">
<h3>📬 Messages reçus</h3>
<span class="ngl-inbox-count">` + fmt.Sprintf("%d", len(msgs)) + ` msg` + func() string { if len(msgs) > 1 { return "s" }; return "" }() + `</span>
</div>
<div class="ngl-msgs" id="anon-msgs">`

		if len(msgs) == 0 {
			content += `<div class="ngl-empty-msgs">Aucun message pour l'instant. Partage ton lien !</div>`
		} else {
			for _, m := range msgs {
				mID := database.DBValue(m["id"])
				mText := database.DBValue(m["message_text"])
				isRead := database.DBBool(m["is_read"])
				mTime := database.DBValue(m["sent_at"])

				unreadDot := ""
				if !isRead {
					unreadDot = `<span class="ngl-dot"></span>`
				}
				readAttr := ""
				if isRead {
					readAttr = " style='opacity:.5'"
				}

				content += fmt.Sprintf(`
<div class="ngl-msg" id="msg-%s"%s>
<div class="ngl-msg-top">
<div class="ngl-msg-unread">%s</div>
<div class="ngl-msg-text">%s</div>
</div>
<div class="ngl-msg-bar">
<span class="ngl-msg-time">%s</span>
<div class="ngl-msg-actions">
<button class="ngl-act-read" onclick="markRead('%s')" %s>✓ Lu</button>
<button class="ngl-act-del" onclick="deleteMsg('%s')">✕ Suppr.</button>
</div>
</div>
</div>`, mID, readAttr, unreadDot, mText, h.timeAgo(mTime),
					mID,
					func() string { if isRead { return "disabled" }; return "" }(),
					mID)
			}
		}

		content += `
</div>
</div>
<div id="ab-reset-msg" class="ngl-toast"></div>
</div>`
	}

	content += `
<style>
.ngl-dash{min-height:100vh;background:linear-gradient(135deg,#0f0c29,#302b63,#24243e);padding:40px 16px;box-sizing:border-box}
.ngl-dash-inner{max-width:600px;margin:0 auto}
.ngl-dash-title{text-align:center;font-size:1.8rem;margin:0 0 24px}
.ngl-empty{background:rgba(255,255,255,.06);backdrop-filter:blur(20px);border:1px solid rgba(255,255,255,.1);border-radius:20px;padding:48px 32px;text-align:center}
.ngl-empty p{color:rgba(255,255,255,.6);margin:0 0 20px;font-size:1.05rem}
.ngl-btn-primary{background:linear-gradient(135deg,#667eea,#764ba2);border:none;color:#fff;padding:12px 36px;border-radius:14px;font-size:1rem;font-weight:600;cursor:pointer;transition:opacity .2s}
.ngl-btn-primary:hover{opacity:.9}
.ngl-share{background:rgba(255,255,255,.06);backdrop-filter:blur(20px);border:1px solid rgba(255,255,255,.1);border-radius:20px;padding:20px;margin-bottom:20px}
.ngl-share-label{font-size:.85rem;color:rgba(255,255,255,.5);margin-bottom:10px}
.ngl-share-row{display:flex;gap:8px;align-items:center;background:rgba(255,255,255,.05);border:1px solid rgba(255,255,255,.1);border-radius:14px;padding:4px 4px 4px 14px}
.ngl-prefix{color:rgba(255,255,255,.4);font-size:.9rem;font-family:monospace;white-space:nowrap}
.ngl-share-row input{flex:1;padding:10px 4px;background:transparent;border:none;color:#fff;font-size:.95rem;outline:none;font-family:monospace}
.ngl-btn-copy{background:rgba(255,255,255,.08);border:none;color:#fff;padding:8px 14px;border-radius:10px;cursor:pointer;font-size:1.1rem;transition:background .2s}
.ngl-btn-copy:hover{background:rgba(255,255,255,.15)}
.ngl-share-actions{display:flex;gap:12px;margin-top:10px}
.ngl-btn-ghost{background:none;border:none;color:rgba(255,255,255,.5);font-size:.8rem;cursor:pointer;padding:4px 0;transition:color .2s}
.ngl-btn-ghost:hover{color:rgba(255,255,255,.8)}
.ngl-inbox{background:rgba(255,255,255,.06);backdrop-filter:blur(20px);border:1px solid rgba(255,255,255,.1);border-radius:20px;padding:20px}
.ngl-inbox-header{display:flex;align-items:center;justify-content:space-between;margin-bottom:16px}
.ngl-inbox-header h3{margin:0;font-size:1.1rem}
.ngl-inbox-count{font-size:.8rem;color:rgba(255,255,255,.4)}
.ngl-msgs{display:flex;flex-direction:column;gap:8px;max-height:500px;overflow-y:auto}
.ngl-msg{background:rgba(255,255,255,.04);border-radius:14px;padding:14px;transition:background .2s}
.ngl-msg:hover{background:rgba(255,255,255,.07)}
.ngl-msg-top{display:flex;gap:8px;align-items:flex-start}
.ngl-dot{width:8px;height:8px;border-radius:50%;background:#667eea;flex-shrink:0;margin-top:6px;display:inline-block}
.ngl-msg-unread{width:8px;flex-shrink:0}
.ngl-msg-text{color:#e2e8f0;line-height:1.5;font-size:.95rem}
.ngl-msg-bar{display:flex;align-items:center;justify-content:space-between;margin-top:10px;padding-left:16px}
.ngl-msg-time{font-size:.75rem;color:rgba(255,255,255,.3)}
.ngl-msg-actions{display:flex;gap:8px}
.ngl-act-read,.ngl-act-del{background:none;border:none;font-size:.78rem;cursor:pointer;padding:4px 10px;border-radius:8px;transition:background .2s}
.ngl-act-read{color:#22c55e}
.ngl-act-read:hover{background:rgba(34,197,94,.1)}
.ngl-act-read:disabled{color:rgba(255,255,255,.2);cursor:default}
.ngl-act-read:disabled:hover{background:none}
.ngl-act-del{color:#ef4444}
.ngl-act-del:hover{background:rgba(239,68,68,.1)}
.ngl-empty-msgs{text-align:center;padding:40px 20px;color:rgba(255,255,255,.4)}
.ngl-toast{text-align:center;margin-top:12px;padding:10px;border-radius:10px;background:rgba(34,197,94,.15);color:#22c55e;font-size:.85rem}
</style>
<script>
function copyAnonLink(){var t=document.getElementById('anon-token');var full=window.location.origin+'/anon/'+t.value;navigator.clipboard.writeText(full).then(function(){var el=document.querySelector('.ngl-toast');el.textContent='Lien copié !';el.style.display='block';setTimeout(function(){el.style.display='none';},2000);}).catch(function(){var i=document.createElement('input');i.value=full;document.body.appendChild(i);i.select();document.execCommand('copy');document.body.removeChild(i);var el=document.querySelector('.ngl-toast');el.textContent='Lien copié !';el.style.display='block';setTimeout(function(){el.style.display='none';},2000);});}

function onAnonTokenChange(){document.getElementById('anon-save-btn').style.display='inline-block';}

function saveAnonToken(){var t=document.getElementById('anon-token').value.trim();if(!t){alert('Le lien ne peut pas être vide');return;}fetch('/api/anon-box',{method:'PATCH',headers:{'Content-Type':'application/json'},body:JSON.stringify({token:t})}).then(function(r){return r.json()}).then(function(d){if(d.success){location.reload();}else alert(d.error||'Erreur');});}

function resetAnonLink(){
if(!confirm('Réinitialiser le lien ? L\'ancien lien sera désactivé.'))return;
fetch('/api/anon-box/reset',{method:'POST'}).then(function(r){return r.json()}).then(function(d){
if(d.success){document.getElementById('anon-token').value=d.token;document.getElementById('anon-save-btn').style.display='none';var el=document.querySelector('.ngl-toast');el.textContent='Lien réinitialisé !';el.style.display='block';setTimeout(function(){el.style.display='none';},3000);}
});
}

function markRead(mid){
fetch('/api/anon-box/messages/'+mid+'/read',{method:'POST'}).then(function(r){return r.json()}).then(function(d){
if(d.success){var el=document.getElementById('msg-'+mid);if(el){el.style.opacity='.5';var dot=el.querySelector('.ngl-dot');if(dot)dot.style.display='none';var btn=el.querySelector('.ngl-act-read');if(btn)btn.disabled=true;}}
});
}

function deleteMsg(mid){
if(!confirm('Supprimer ce message ?'))return;
fetch('/api/anon-box/messages/'+mid+'/delete',{method:'POST'}).then(function(r){return r.json()}).then(function(d){
if(d.success){var el=document.getElementById('msg-'+mid);if(el)el.remove();}
});
}
</script>`

	return renderPage(c, "Messages Anonymes", content)
}

func (h *Handler) AnonBoxCreatePost(c *fiber.Ctx) error {
	return h.AnonBoxCreate(c)
}

// ============================================================
// ANON BOX — PUBLIC PAGE
// ============================================================

func (h *Handler) AnonBoxPublic(c *fiber.Ctx) error {
	token := c.Params("token")
	log.Printf("[AnonBoxPublic] token=%s", token)

	boxBody, _ := h.db.Select("anon_boxes",
		fmt.Sprintf("token=eq.%s&active=eq.true&select=*", token), true)
	log.Printf("[AnonBoxPublic] select body=%s", string(boxBody))
	var boxes []map[string]interface{}
	json.Unmarshal(boxBody, &boxes)
	if len(boxes) == 0 {
		log.Printf("[AnonBoxPublic] box not found for token=%s", token)
		return c.Status(404).SendString("Boîte introuvable ou désactivée")
	}
	box := boxes[0]
	boxID := database.DBValue(box["id"])
	ownerID := database.DBValue(box["owner_id"])

	ownerBody, _ := h.db.Select("user_profiles",
		fmt.Sprintf("id=eq.%s&select=nickname,username,avatar_url", ownerID), false)
	var owners []map[string]interface{}
	json.Unmarshal(ownerBody, &owners)
	ownerName := "Quelqu'un"
	ownerAvatar := ""
	if len(owners) > 0 {
		if nn, ok := owners[0]["nickname"].(string); ok && nn != "" {
			ownerName = nn
		} else if un, ok := owners[0]["username"].(string); ok {
			ownerName = un
		}
		if av, ok := owners[0]["avatar_url"].(string); ok {
			ownerAvatar = av
		}
	}
	avatarHTML := `<div class="ngl-avatar">` + ownerName[:1] + `</div>`
	if ownerAvatar != "" {
		avatarHTML = `<img class="ngl-avatar-img" src="` + htmlAttr(ownerAvatar) + `" alt="">`
	}

	content := `
<div class="ngl-page">
<div class="ngl-card">
` + avatarHTML + `
<h2 class="ngl-name">` + ownerName + `</h2>
<p class="ngl-subtitle">Envoie-moi un message anonyme 💬</p>

<form id="anon-form" onsubmit="sendAnonMsg(event)">
<div class="ngl-input-wrap">
<textarea id="anon-msg-text" placeholder="Écris ton message ici..." maxlength="500" rows="3"></textarea>
</div>
<div class="ngl-footer">
<span id="char-count">0/500</span>
<button type="submit" id="send-btn">Envoyer</button>
</div>
</form>
<div id="anon-confirm" style="display:none;text-align:center;padding:20px">
<div class="ngl-check">✓</div>
<h3 style="margin:12px 0 4px;font-size:1.3rem">Message envoyé !</h3>
<p style="color:rgba(255,255,255,.5);margin:0 0 4px">Merci pour ton message.</p>
<p style="color:rgba(255,255,255,.4);font-size:.85rem;margin:0 0 20px">Tu veux toi aussi recevoir des messages anonymes ? <a href="/register" style="color:#667eea;text-decoration:none;font-weight:600">Inscris-toi !</a></p>
<button onclick="resetForm()" class="ngl-btn-outline">Envoyer un autre</button>
</div>
</div>
</div>
<style>
.ngl-page{min-height:100vh;display:flex;align-items:center;justify-content:center;padding:20px;background:linear-gradient(135deg,#0f0c29,#302b63,#24243e);box-sizing:border-box}
.ngl-card{background:rgba(255,255,255,.06);backdrop-filter:blur(20px);-webkit-backdrop-filter:blur(20px);border:1px solid rgba(255,255,255,.1);border-radius:24px;padding:40px 32px;width:100%;max-width:420px;text-align:center;box-shadow:0 25px 50px -12px rgba(0,0,0,.5)}
.ngl-avatar,.ngl-avatar-img{width:72px;height:72px;border-radius:50%;margin:0 auto 12px;box-shadow:0 8px 24px rgba(102,126,234,.3)}
.ngl-avatar{background:linear-gradient(135deg,#667eea,#764ba2);display:flex;align-items:center;justify-content:center;font-size:1.8rem;font-weight:700;color:#fff}
.ngl-avatar-img{object-fit:cover;display:block}
.ngl-name{margin:0;font-size:1.5rem;font-weight:700}
.ngl-subtitle{color:rgba(255,255,255,.5);margin:4px 0 24px;font-size:.95rem}
.ngl-input-wrap{background:rgba(255,255,255,.05);border:1px solid rgba(255,255,255,.1);border-radius:16px;padding:4px;transition:border-color .2s}
.ngl-input-wrap:focus-within{border-color:rgba(102,126,234,.6)}
.ngl-input-wrap textarea{width:100%;padding:14px;background:transparent;border:none;color:#fff;font-size:1rem;resize:vertical;outline:none;font-family:inherit;box-sizing:border-box;min-height:80px}
.ngl-input-wrap textarea::placeholder{color:rgba(255,255,255,.3)}
.ngl-footer{display:flex;justify-content:space-between;align-items:center;margin-top:14px;gap:12px}
.ngl-footer span{font-size:.8rem;color:rgba(255,255,255,.35)}
.ngl-footer button{background:linear-gradient(135deg,#667eea,#764ba2);border:none;color:#fff;padding:10px 28px;border-radius:12px;font-size:.95rem;font-weight:600;cursor:pointer;transition:transform .15s,opacity .2s}
.ngl-footer button:hover{opacity:.9}
.ngl-footer button:active{transform:scale(.97)}
.ngl-footer button:disabled{opacity:.5;cursor:default}
.ngl-check{width:56px;height:56px;border-radius:50%;background:linear-gradient(135deg,#22c55e,#16a34a);display:flex;align-items:center;justify-content:center;font-size:1.5rem;color:#fff;margin:0 auto;box-shadow:0 8px 24px rgba(34,197,94,.3)}
.ngl-btn-outline{background:transparent;border:1px solid rgba(255,255,255,.15);color:#fff;padding:10px 28px;border-radius:12px;font-size:.95rem;cursor:pointer;transition:background .2s}
.ngl-btn-outline:hover{background:rgba(255,255,255,.05)}
</style>
<script>
var BOX_ID='` + boxID + `';
var msgText=document.getElementById('anon-msg-text');
msgText.addEventListener('input',function(){document.getElementById('char-count').textContent=this.value.length+'/500';});

function sendAnonMsg(e){
e.preventDefault();
var text=msgText.value.trim();
if(!text)return;
var btn=document.getElementById('send-btn');
btn.disabled=true;btn.textContent='Envoi...';
fetch('/api/anon-box/'+BOX_ID+'/send',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({message:text})}).then(function(r){return r.json()}).then(function(d){
if(d.success){
document.getElementById('anon-form').style.display='none';
document.getElementById('anon-confirm').style.display='block';
}else{alert(d.error||"Erreur");btn.disabled=false;btn.textContent='Envoyer';}
});
}
function resetForm(){
msgText.value='';document.getElementById('char-count').textContent='0/500';
document.getElementById('anon-form').style.display='block';document.getElementById('anon-confirm').style.display='none';
document.getElementById('send-btn').disabled=false;document.getElementById('send-btn').textContent='Envoyer';
}
</script>`

	return renderPage(c, "Message anonyme", content)
}

// ============================================================
// ANON BOX — API
// ============================================================

func (h *Handler) APIAnonBoxSend(c *fiber.Ctx) error {
	boxID := c.Params("id")
	log.Printf("[APIAnonBoxSend] boxID=%s", boxID)

	boxBody, _ := h.db.Select("anon_boxes",
		fmt.Sprintf("id=eq.%s&active=eq.true&select=id,owner_id", boxID), false)
	log.Printf("[APIAnonBoxSend] box select body=%s", string(boxBody))
	var boxes []map[string]interface{}
	json.Unmarshal(boxBody, &boxes)
	if len(boxes) == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Boîte introuvable"})
	}
	ownerID := database.DBValue(boxes[0]["owner_id"])

	type Request struct {
		Message string `json:"message"`
	}
	var req Request
	c.BodyParser(&req)

	if req.Message == "" || len(req.Message) > 500 {
		return c.Status(400).JSON(fiber.Map{"error": "Message invalide (max 500 caractères)"})
	}

	senderIP := c.IP()
	ipHash := fmt.Sprintf("%x", sha256.Sum256([]byte(senderIP)))
	rateKey := fmt.Sprintf("anon_rate:%s:%s", boxID, ipHash)

	rateBody, _ := h.db.Select("anon_rate_limits",
		fmt.Sprintf("key=eq.%s&select=id,expires_at", rateKey), false)
	var rateRows []map[string]interface{}
	json.Unmarshal(rateBody, &rateRows)

	now := time.Now()
	if len(rateRows) > 0 {
		expiresStr := database.DBValue(rateRows[0]["expires_at"])
		expires, err := time.Parse(time.RFC3339, expiresStr)
		if err == nil && now.Before(expires) {
			return c.Status(429).JSON(fiber.Map{"error": "Vous avez atteint la limite d'envoi. Réessayez plus tard."})
		}
		h.db.Delete("anon_rate_limits", fmt.Sprintf("key=eq.%s", rateKey), false)
	}

	rateData, _ := json.Marshal(map[string]interface{}{
		"key":        rateKey,
		"expires_at": now.Add(1 * time.Hour).Format(time.RFC3339),
	})
	h.db.Insert("anon_rate_limits", rateData, false)

	msgData, _ := json.Marshal(map[string]interface{}{
		"box_id":       boxID,
		"message_text": req.Message,
		"sender_ip":    ipHash,
		"is_read":      false,
	})
	log.Printf("[APIAnonBoxSend] inserting message: %s", string(msgData))
	msgBody, msgErr := h.db.Insert("anon_messages", msgData, false)
	log.Printf("[APIAnonBoxSend] insert result: body=%s err=%v", string(msgBody), msgErr)

	notifData, _ := json.Marshal(map[string]interface{}{
		"user_id":    ownerID,
		"type":       "anon_message",
		"title":      "Nouveau message anonyme !",
		"message":    "Quelqu'un vous a envoyé un message anonyme !",
		"linked_id":  boxID,
		"is_read":    false,
	})
	h.db.Insert("notifications", notifData, true)

	log.Printf("[APIAnonBoxSend] success for box=%s", boxID)
	return c.JSON(fiber.Map{"success": true})
}

func (h *Handler) APIAnonBoxMarkRead(c *fiber.Ctx) error {
	user := h.getUserFromSession(c)
	msgID := c.Params("mid")

	msgBody, _ := h.db.Select("anon_messages",
		fmt.Sprintf("id=eq.%s&select=id,box_id", msgID), true)
	var msgs []map[string]interface{}
	json.Unmarshal(msgBody, &msgs)
	if len(msgs) == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Message introuvable"})
	}

	boxID := database.DBValue(msgs[0]["box_id"])

	boxBody, _ := h.db.Select("anon_boxes",
		fmt.Sprintf("id=eq.%s&owner_id=eq.%s&select=id", boxID, user.ID), true)
	var boxes []map[string]interface{}
	json.Unmarshal(boxBody, &boxes)
	if len(boxes) == 0 {
		return c.Status(403).JSON(fiber.Map{"error": "Non autorisé"})
	}

	h.db.Update("anon_messages",
		fmt.Sprintf("id=eq.%s", msgID),
		[]byte(`{"is_read":true}`), true)

	return c.JSON(fiber.Map{"success": true})
}

func (h *Handler) APIAnonBoxDeleteMsg(c *fiber.Ctx) error {
	user := h.getUserFromSession(c)
	msgID := c.Params("mid")

	msgBody, _ := h.db.Select("anon_messages",
		fmt.Sprintf("id=eq.%s&select=id,box_id", msgID), true)
	var msgs []map[string]interface{}
	json.Unmarshal(msgBody, &msgs)
	if len(msgs) == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Message introuvable"})
	}

	boxID := database.DBValue(msgs[0]["box_id"])

	boxBody, _ := h.db.Select("anon_boxes",
		fmt.Sprintf("id=eq.%s&owner_id=eq.%s&select=id", boxID, user.ID), true)
	var boxes []map[string]interface{}
	json.Unmarshal(boxBody, &boxes)
	if len(boxes) == 0 {
		return c.Status(403).JSON(fiber.Map{"error": "Non autorisé"})
	}

	h.db.Delete("anon_messages", fmt.Sprintf("id=eq.%s", msgID), true)

	return c.JSON(fiber.Map{"success": true})
}

func (h *Handler) APIAnonBoxReset(c *fiber.Ctx) error {
	user := h.getUserFromSession(c)

	boxBody, _ := h.db.Select("anon_boxes",
		fmt.Sprintf("owner_id=eq.%s&active=eq.true&select=id,token", user.ID), true)
	var boxes []map[string]interface{}
	json.Unmarshal(boxBody, &boxes)
	if len(boxes) == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Boîte introuvable"})
	}

	boxID := database.DBValue(boxes[0]["id"])

	h.db.Update("anon_boxes",
		fmt.Sprintf("id=eq.%s", boxID),
		[]byte(`{"active":false}`), true)

	newToken := generateToken(8)
	data, _ := json.Marshal(map[string]interface{}{
		"owner_id": user.ID,
		"token":    newToken,
		"active":   true,
	})
	h.db.Insert("anon_boxes", data, true)

	return c.JSON(fiber.Map{"success": true, "token": newToken})
}

func (h *Handler) APIAnonBoxUpdate(c *fiber.Ctx) error {
	user := h.getUserFromSession(c)

	boxBody, _ := h.db.Select("anon_boxes",
		fmt.Sprintf("owner_id=eq.%s&active=eq.true&select=id,token", user.ID), true)
	var boxes []map[string]interface{}
	json.Unmarshal(boxBody, &boxes)
	if len(boxes) == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Boîte introuvable"})
	}

	boxID := database.DBValue(boxes[0]["id"])

	type Request struct {
		Token string `json:"token"`
	}
	var req Request
	if err := c.BodyParser(&req); err != nil || req.Token == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Token invalide"})
	}

	dupBody, _ := h.db.Select("anon_boxes",
		fmt.Sprintf("token=eq.%s&id=neq.%s&select=id,owner_id,active", req.Token, boxID), true)
	var dup []map[string]interface{}
	json.Unmarshal(dupBody, &dup)
	if len(dup) > 0 {
		d := dup[0]
		if database.DBBool(d["active"]) {
			return c.Status(400).JSON(fiber.Map{"error": "Ce lien est déjà utilisé"})
		}
		dupID := database.DBValue(d["id"])
		h.db.Delete("anon_boxes", fmt.Sprintf("id=eq.%s", dupID), true)
	}

	req.Token = strings.TrimSpace(req.Token)
	updateData, _ := json.Marshal(map[string]string{"token": req.Token})
	if _, err := h.db.Update("anon_boxes",
		fmt.Sprintf("id=eq.%s", boxID), updateData, true); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Erreur lors de la sauvegarde"})
	}

	return c.JSON(fiber.Map{"success": true})
}
