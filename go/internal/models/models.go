package models

import "time"

type User struct {
	ID             string    `json:"id"`
	Email          string    `json:"email"`
	Username       string    `json:"username"`
	Nickname       *string   `json:"nickname"`
	AvatarURL      *string   `json:"avatar_url"`
	Bio            *string   `json:"bio"`
	Country        *string   `json:"country"`
	Phone          *string   `json:"phone"`
	FavoriteAnime  *string   `json:"favorite_anime"`
	XP             int       `json:"xp"`
	Level          int       `json:"level"`
	Rank           string    `json:"rank"`
	IsPremium      bool      `json:"is_premium"`
	IsAdmin        bool      `json:"is_admin"`
	CanCreateQuiz  bool      `json:"can_create_quiz"`
	CurrentStreak  int       `json:"current_streak"`
	LongestStreak  int       `json:"longest_streak"`
	LastLoginDate  *string   `json:"last_login_date"`
	TotalXP        int       `json:"total_xp"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type Quiz struct {
	ID              string    `json:"id"`
	CreatorID       string    `json:"creator_id"`
	Title           string    `json:"title"`
	Description     *string   `json:"description"`
	ThumbnailURL    *string   `json:"thumbnail_url"`
	Category        string    `json:"category"`
	Subcategory     string    `json:"subcategory"`
	Series          string    `json:"series"`
	QuizType        string    `json:"quiz_type"`
	Status          string    `json:"status"`
	QuestionCount   int       `json:"question_count"`
	PlayCount       int       `json:"play_count"`
	IsVisible       bool      `json:"is_visible"`
	StartsAt        *time.Time `json:"starts_at"`
	EndsAt          *time.Time `json:"ends_at"`
	DurationSeconds *int      `json:"duration_seconds"`
	DurationMode    *string   `json:"duration_mode"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	Creator         *User     `json:"creator,omitempty"`
}

type Question struct {
	ID               string    `json:"id"`
	QuizID           string    `json:"quiz_id"`
	QuestionText     string    `json:"question_text"`
	QuestionType     string    `json:"question_type"`
	MediaURL         *string   `json:"media_url"`
	TimeLimitSeconds int       `json:"time_limit_seconds"`
	OrderIndex       int       `json:"order_index"`
	Answers          []Answer  `json:"answers,omitempty"`
}

type Answer struct {
	ID         string `json:"id"`
	QuestionID string `json:"question_id"`
	AnswerText string `json:"answer_text"`
	IsCorrect  bool   `json:"is_correct"`
	OrderIndex int    `json:"order_index"`
}

type Friendship struct {
	ID          string    `json:"id"`
	RequesterID string    `json:"requester_id"`
	AddresseeID string    `json:"addressee_id"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Friend      *User     `json:"friend,omitempty"`
}

type Message struct {
	ID             string    `json:"id"`
	ConversationID string    `json:"conversation_id"`
	SenderID       string    `json:"sender_id"`
	Content        string    `json:"content"`
	IsRead         bool      `json:"is_read"`
	CreatedAt      time.Time `json:"created_at"`
	Sender         *User     `json:"sender,omitempty"`
}

type Conversation struct {
	ID           string    `json:"id"`
	User1ID      string    `json:"user1_id"`
	User2ID      string    `json:"user2_id"`
	LastMessageAt *time.Time `json:"last_message_at"`
	CreatedAt    time.Time `json:"created_at"`
	OtherUser    *User     `json:"other_user,omitempty"`
	LastMessage  *Message  `json:"last_message,omitempty"`
	UnreadCount  int       `json:"unread_count"`
}

type GameSession struct {
	ID            string     `json:"id"`
	UserID        string     `json:"user_id"`
	QuizID        string     `json:"quiz_id"`
	StartedAt     time.Time  `json:"started_at"`
	CompletedAt   *time.Time `json:"completed_at"`
	Score         int        `json:"score"`
	CorrectCount  int        `json:"correct_count"`
	TotalQuestions int       `json:"total_questions"`
	AccuracyRate  float64    `json:"accuracy_rate"`
	IsPerfect     bool       `json:"is_perfect"`
	TimeTakenMs   *int       `json:"time_taken_ms"`
	CreatedAt     time.Time  `json:"created_at"`
}

type ChallengeSession struct {
	ID             string     `json:"id"`
	QuizID         string     `json:"quiz_id"`
	CreatorID      string     `json:"creator_id"`
	MinPlayers     int        `json:"min_players"`
	InviteExpiresAt time.Time `json:"invite_expires_at"`
	Status         string     `json:"status"`
	WinnerID       *string    `json:"winner_id"`
	TotalXPPool    int        `json:"total_xp_pool"`
	StartedAt      *time.Time `json:"started_at"`
	CompletedAt    *time.Time `json:"completed_at"`
	CreatedAt      time.Time  `json:"created_at"`
	Quiz           *Quiz      `json:"quiz,omitempty"`
	Creator        *User      `json:"creator,omitempty"`
	Participants   []ChallengeParticipant `json:"participants,omitempty"`
}

type ChallengeParticipant struct {
	ID          string  `json:"id"`
	SessionID   string  `json:"session_id"`
	UserID      string  `json:"user_id"`
	XPBet       int     `json:"xp_bet"`
	Status      string  `json:"status"`
	Score       int     `json:"score"`
	XPWon       int     `json:"xp_won"`
	XPLost      int     `json:"xp_lost"`
	User        *User   `json:"user,omitempty"`
}

type Notification struct {
	ID        string                 `json:"id"`
	UserID    string                 `json:"user_id"`
	Type      string                 `json:"type"`
	Title     string                 `json:"title"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data"`
	IsRead    bool                   `json:"is_read"`
	CreatedAt time.Time              `json:"created_at"`
}
