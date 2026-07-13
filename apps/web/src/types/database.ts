// ============================================================
// TYPES SUPABASE DATABASE - Généré manuellement
// ============================================================

export type Database = {
  public: {
    Tables: {
      user_profiles: {
        Row: {
          id: string
          email: string
          username: string
          avatar_url: string | null
          bio: string | null
          country: string | null
          phone: string | null
          favorite_anime: string | null
          xp: number
          level: number
          rank: string
          is_premium: boolean
          created_at: string
          updated_at: string
        }
        Insert: {
          id: string
          email: string
          username: string
          avatar_url?: string | null
          bio?: string | null
          country?: string | null
          phone?: string | null
          favorite_anime?: string | null
          xp?: number
          level?: number
          rank?: string
          is_premium?: boolean
          created_at?: string
          updated_at?: string
        }
        Update: {
          id?: string
          email?: string
          username?: string
          avatar_url?: string | null
          bio?: string | null
          country?: string | null
          phone?: string | null
          favorite_anime?: string | null
          xp?: number
          level?: number
          rank?: string
          is_premium?: boolean
          created_at?: string
          updated_at?: string
        }
      }
      user_stats: {
        Row: {
          user_id: string
          quizzes_played: number
          quizzes_created: number
          total_correct_answers: number
          total_answers: number
          accuracy_rate: number
          best_score_ever: number
          monthly_rank: number | null
          global_rank: number | null
          updated_at: string
        }
        Insert: {
          user_id: string
          quizzes_played?: number
          quizzes_created?: number
          total_correct_answers?: number
          total_answers?: number
          accuracy_rate?: number
          best_score_ever?: number
          monthly_rank?: number | null
          global_rank?: number | null
          updated_at?: string
        }
        Update: {
          user_id?: string
          quizzes_played?: number
          quizzes_created?: number
          total_correct_answers?: number
          total_answers?: number
          accuracy_rate?: number
          best_score_ever?: number
          monthly_rank?: number | null
          global_rank?: number | null
          updated_at?: string
        }
      }
      quizzes: {
        Row: {
          id: string
          creator_id: string
          title: string
          description: string | null
          thumbnail_url: string | null
          thumbnail_public_id: string | null
          category: string
          subcategory: string
          series: string
          quiz_type: string
          status: string
          question_count: number
          play_count: number
          total_reports: number
          is_visible: boolean
          event_start_at: string | null
          event_end_at: string | null
          created_at: string
          updated_at: string
        }
        Insert: {
          id?: string
          creator_id: string
          title: string
          description?: string | null
          thumbnail_url?: string | null
          thumbnail_public_id?: string | null
          category: string
          subcategory: string
          series: string
          quiz_type?: string
          status?: string
          question_count?: number
          play_count?: number
          total_reports?: number
          is_visible?: boolean
          event_start_at?: string | null
          event_end_at?: string | null
          created_at?: string
          updated_at?: string
        }
        Update: {
          id?: string
          creator_id?: string
          title?: string
          description?: string | null
          thumbnail_url?: string | null
          thumbnail_public_id?: string | null
          category?: string
          subcategory?: string
          series?: string
          quiz_type?: string
          status?: string
          question_count?: number
          play_count?: number
          total_reports?: number
          is_visible?: boolean
          event_start_at?: string | null
          event_end_at?: string | null
          created_at?: string
          updated_at?: string
        }
      }
      questions: {
        Row: {
          id: string
          quiz_id: string
          question_text: string
          question_type: string
          media_url: string | null
          media_public_id: string | null
          time_limit_seconds: number
          order_index: number
          created_at: string
        }
        Insert: {
          id?: string
          quiz_id: string
          question_text: string
          question_type: string
          media_url?: string | null
          media_public_id?: string | null
          time_limit_seconds?: number
          order_index: number
          created_at?: string
        }
        Update: {
          id?: string
          quiz_id?: string
          question_text?: string
          question_type?: string
          media_url?: string | null
          media_public_id?: string | null
          time_limit_seconds?: number
          order_index?: number
          created_at?: string
        }
      }
      answers: {
        Row: {
          id: string
          question_id: string
          answer_text: string
          is_correct: boolean
          order_index: number
          created_at: string
        }
        Insert: {
          id?: string
          question_id: string
          answer_text: string
          is_correct?: boolean
          order_index: number
          created_at?: string
        }
        Update: {
          id?: string
          question_id?: string
          answer_text?: string
          is_correct?: boolean
          order_index?: number
          created_at?: string
        }
      }
      game_sessions: {
        Row: {
          id: string
          user_id: string
          quiz_id: string
          started_at: string
          completed_at: string | null
          score: number
          correct_count: number
          total_questions: number
          accuracy_rate: number
          is_perfect: boolean
          time_taken_ms: number | null
          created_at: string
        }
        Insert: {
          id?: string
          user_id: string
          quiz_id: string
          started_at?: string
          completed_at?: string | null
          score?: number
          correct_count?: number
          total_questions: number
          accuracy_rate?: number
          is_perfect?: boolean
          time_taken_ms?: number | null
          created_at?: string
        }
        Update: {
          id?: string
          user_id?: string
          quiz_id?: string
          started_at?: string
          completed_at?: string | null
          score?: number
          correct_count?: number
          total_questions?: number
          accuracy_rate?: number
          is_perfect?: boolean
          time_taken_ms?: number | null
          created_at?: string
        }
      }
      player_answers: {
        Row: {
          id: string
          session_id: string
          question_id: string
          answer_id: string | null
          is_correct: boolean
          time_taken_ms: number
          points_earned: number
          created_at: string
        }
        Insert: {
          id?: string
          session_id: string
          question_id: string
          answer_id?: string | null
          is_correct: boolean
          time_taken_ms: number
          points_earned?: number
          created_at?: string
        }
        Update: {
          id?: string
          session_id?: string
          question_id?: string
          answer_id?: string | null
          is_correct?: boolean
          time_taken_ms?: number
          points_earned?: number
          created_at?: string
        }
      }
      friendships: {
        Row: {
          id: string
          requester_id: string
          addressee_id: string
          status: string
          created_at: string
          updated_at: string
        }
        Insert: {
          id?: string
          requester_id: string
          addressee_id: string
          status?: string
          created_at?: string
          updated_at?: string
        }
        Update: {
          id?: string
          requester_id?: string
          addressee_id?: string
          status?: string
          created_at?: string
          updated_at?: string
        }
      }
      badges: {
        Row: {
          id: string
          slug: string
          name: string
          description: string
          icon_url: string | null
          condition_type: string
          condition_value: number
          is_rare: boolean
          created_at: string
        }
        Insert: {
          id?: string
          slug: string
          name: string
          description: string
          icon_url?: string | null
          condition_type: string
          condition_value: number
          is_rare?: boolean
          created_at?: string
        }
        Update: {
          id?: string
          slug?: string
          name?: string
          description?: string
          icon_url?: string | null
          condition_type?: string
          condition_value?: number
          is_rare?: boolean
          created_at?: string
        }
      }
      user_badges: {
        Row: {
          id: string
          user_id: string
          badge_id: string
          earned_at: string
        }
        Insert: {
          id?: string
          user_id: string
          badge_id: string
          earned_at?: string
        }
        Update: {
          id?: string
          user_id?: string
          badge_id?: string
          earned_at?: string
        }
      }
      reports: {
        Row: {
          id: string
          reporter_id: string
          quiz_id: string
          reason: string
          description: string | null
          status: string
          created_at: string
        }
        Insert: {
          id?: string
          reporter_id: string
          quiz_id: string
          reason: string
          description?: string | null
          status?: string
          created_at?: string
        }
        Update: {
          id?: string
          reporter_id?: string
          quiz_id?: string
          reason?: string
          description?: string | null
          status?: string
          created_at?: string
        }
      }
      leaderboard_monthly: {
        Row: {
          id: string
          user_id: string
          month_year: string
          score: number
          rank_position: number | null
          created_at: string
        }
        Insert: {
          id?: string
          user_id: string
          month_year: string
          score?: number
          rank_position?: number | null
          created_at?: string
        }
        Update: {
          id?: string
          user_id?: string
          month_year?: string
          score?: number
          rank_position?: number | null
          created_at?: string
        }
      }
      notifications: {
        Row: {
          id: string
          user_id: string
          type: string
          title: string
          message: string
          data: Record<string, unknown> | null
          is_read: boolean
          created_at: string
        }
        Insert: {
          id?: string
          user_id: string
          type: string
          title: string
          message: string
          data?: Record<string, unknown> | null
          is_read?: boolean
          created_at?: string
        }
        Update: {
          id?: string
          user_id?: string
          type?: string
          title?: string
          message?: string
          data?: Record<string, unknown> | null
          is_read?: boolean
          created_at?: string
        }
      }
    }
    Views: {
      [_ in never]: never
    }
    Functions: {
      get_global_leaderboard: {
        Args: { limit_count?: number }
        Returns: {
          rank: number
          user_id: string
          username: string
          avatar_url: string | null
          user_rank: string
          xp: number
          quizzes_played: number
        }[]
      }
      get_monthly_leaderboard: {
        Args: { year_month: string; limit_count?: number }
        Returns: {
          rank: number
          user_id: string
          username: string
          avatar_url: string | null
          user_rank: string
          score: number
          quiz_count: number
        }[]
      }
      get_quiz_leaderboard: {
        Args: { quiz_id: string }
        Returns: {
          rank: number
          user_id: string
          username: string
          avatar_url: string | null
          user_rank: string
          score: number
          accuracy_rate: number
          time_taken_ms: number
        }[]
      }
      get_series_leaderboard: {
        Args: { series_name: string; limit_count?: number }
        Returns: {
          rank: number
          user_id: string
          username: string
          avatar_url: string | null
          user_rank: string
          total_score: number
          quiz_count: number
        }[]
      }
      check_and_award_badges: {
        Args: { target_user_id: string }
        Returns: {
          badge_id: string
          badge_name: string
        }[]
      }
    }
    Enums: {
      [_ in never]: never
    }
  }
}
