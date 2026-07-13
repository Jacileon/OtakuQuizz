package handlers

import (
	"fmt"
	"html/template"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"otaku-quiz-africa/pkg/database"
)

type DashboardStats struct {
	TotalPlayed    int
	BestScore      int
	Accuracy       float64
	MonthlyRank    int
	TotalXP        int
	CurrentStreak  int
}

func (h *Handler) getUserFromSession(c *fiber.Ctx) *UserProfile {
	user := c.Locals("user")
	if user == nil {
		return nil
	}
	u, ok := user.(*UserProfile)
	if !ok {
		return nil
	}
	return u
}

func renderWithContent(c *fiber.Ctx, title, content string) error {
	data := fiber.Map{
		"Title":   title,
		"Content": template.HTML(content),
	}
	if u := c.Locals("user"); u != nil {
		data["User"] = u
		if user, ok := u.(*UserProfile); ok {
			badge := challengeStatsBadge(user.ChallengesPlayed, user.ChallengesWon)
			if badge != "" {
				data["StatsBadge"] = template.HTML(badge)
			}
		}
	}
	return c.Render("layouts/main", data)
}

func formatNumber(n int) string {
	if n >= 1000000 {
		return fmt.Sprintf("%.1fM", float64(n)/1000000)
	}
	if n >= 1000 {
		return fmt.Sprintf("%.1fk", float64(n)/1000)
	}
	return fmt.Sprintf("%d", n)
}

func (h *Handler) calcXPPercent(xp int, rank string) int {
	ranks := h.loadRanks()
	return calcXPPercentFromRanks(xp, ranks)
}

func (h *Handler) renderEventCard(event map[string]interface{}, status string) string {
	title, _ := event["title"].(string)
	id, _ := event["id"].(string)
	series, _ := event["series"].(string)
	qCount := database.DBInt(event["question_count"])

	badge := ""
	if status == "live" {
		badge = `<span class="badge-live">🔴 EN DIRECT</span>`
	} else {
		badge = `<span class="badge-upcoming">📅 À VENIR</span>`
	}

	return fmt.Sprintf(`
<div class="ec-card">
    <div class="ec-header">
        <span class="ec-series">%s</span>
        %s
    </div>
    <h3 class="ec-title">%s</h3>
    <div class="ec-meta">%d questions</div>
    <a href="/quiz/%s" class="btn-sm">Voir →</a>
</div>`, series, badge, title, qCount, id)
}

func (h *Handler) renderEventCardLarge(event map[string]interface{}) string {
	title, _ := event["title"].(string)
	id, _ := event["id"].(string)
	description := database.DBValue(event["description"])
	series, _ := event["series"].(string)
	qCount := database.DBInt(event["question_count"])

	return fmt.Sprintf(`
<div class="event-card-large">
    <div class="ecl-header">
        <span class="ecl-series">📺 %s</span>
        <a href="/quiz/%s" class="btn-primary btn-sm">Jouer</a>
    </div>
    <h3>%s</h3>
    <p class="text-muted text-sm">%s</p>
    <div class="ecl-meta">%d questions</div>
</div>`, series, id, title, description, qCount)
}

func (h *Handler) renderQuizCard(q map[string]interface{}) string {
	title, _ := q["title"].(string)
	id, _ := q["id"].(string)
	series, _ := q["series"].(string)
	category, _ := q["category"].(string)
	qCount := database.DBInt(q["question_count"])
	pCount := database.DBInt(q["play_count"])
	qType, _ := q["quiz_type"].(string)

	badge := ""
	if qType == "official" {
		badge = `<span class="badge-official">Officiel</span>`
	}

	return fmt.Sprintf(`
<div class="quiz-card">
    <div class="qc-header">
        <span class="qc-series">%s</span>
        %s
    </div>
    <h3 class="qc-title">%s</h3>
    <div class="qc-meta">%s &bull; %d questions &bull; %d plays</div>
    <div class="qc-actions">
        <a href="/quiz/%s/play" class="btn-primary btn-sm">🎮 Jouer</a>
        <a href="/quiz/%s" class="btn-ghost btn-sm">Détails</a>
    </div>
</div>`, series, badge, title, category, qCount, pCount, id, id)
}

func (h *Handler) timeAgo(dateStr string) string {
	t, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return "récemment"
	}
	diff := time.Since(t)

	if diff < time.Minute {
		return "à l'instant"
	}
	if diff < time.Hour {
		mins := int(diff.Minutes())
		if mins == 1 {
			return "il y a 1 minute"
		}
		return fmt.Sprintf("il y a %d minutes", mins)
	}
	if diff < 24*time.Hour {
		hours := int(diff.Hours())
		if hours == 1 {
			return "il y a 1 heure"
		}
		return fmt.Sprintf("il y a %d heures", hours)
	}
	days := int(diff.Hours() / 24)
	if days == 1 {
		return "hier"
	}
	return fmt.Sprintf("il y a %d jours", days)
}

func getRankColor(rank string) string {
	colors := map[string]string{
		"F": "#888888", "E": "#4CAF50", "D": "#2196F3", "C": "#9C27B0",
		"B": "#FF9800", "A": "#F44336", "S": "#FFD700", "S+": "#FFA500",
		"SS": "#FF69B4", "SSS": "#00FFFF", "Légende": "#FF0080",
	}
	if c, ok := colors[rank]; ok {
		return c
	}
	return "#888888"
}

func getInitials(name string) string {
	if len(name) == 0 {
		return "?"
	}
	return strings.ToUpper(string(name[0]))
}

func getDisplayName(profile *UserProfile) string {
	if profile.Nickname != nil && *profile.Nickname != "" {
		return *profile.Nickname
	}
	return profile.Username
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func htmlAttr(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}

func challengeStatsBadge(played, won int) string {
	if played == 0 {
		return ""
	}
	return fmt.Sprintf(`<span style="display:inline-block;padding:1px 6px;border-radius:4px;background:#6366f122;color:#a78bfa;font-size:.7rem;font-weight:600;margin-left:4px;vertical-align:middle">🏅 %d/%d</span>`, won, played)
}

func challengeStatsBadgeMap(prof map[string]interface{}) string {
	if prof == nil {
		return ""
	}
	played := database.DBInt(prof["challenges_played"])
	won := database.DBInt(prof["challenges_won"])
	return challengeStatsBadge(played, won)
}
