import { createClient } from '@/lib/supabase/server';

// Barème XP par type de question
const XP_BY_QUESTION_TYPE: Record<string, number> = {
  text: 1,
  true_false: 1,
  image: 2,
  gif: 2,
  audio: 3,
  character_guess: 5,
  impostor: 5,
};

export async function calculateQuestionXP(
  userId: string,
  quizId: string,
  questionId: string,
  questionType: string,
  isCorrect: boolean,
  attemptNumber: number
): Promise<number> {
  if (!isCorrect) return 0;

  const baseXP = XP_BY_QUESTION_TYPE[questionType] || 1;

  try {
    // Vérifier si la question a déjà été répondue correctement
    const supabase = await createClient();
    const { data: previousCorrect } = await supabase
      .from('user_question_attempts')
      .select('id')
      .eq('user_id', userId)
      .eq('quiz_id', quizId)
      .eq('question_id', questionId)
      .eq('is_correct', true)
      .limit(1);

    if (previousCorrect && previousCorrect.length > 0) {
      return 0; // Déjà gagné, pas d'XP
    }

    // Calculer la dégressivité
    if (attemptNumber === 1) {
      return baseXP;
    } else if (attemptNumber === 2) {
      return Math.max(1, Math.round(baseXP / 2));
    } else if (attemptNumber === 3) {
      return Math.max(1, Math.round(baseXP / 4));
    } else {
      return 0;
    }
  } catch (error) {
    console.error('Erreur calculateQuestionXP:', error);
    return baseXP; // En cas d'erreur, donner l'XP de base
  }
}

export async function getUserAttemptNumber(
  userId: string,
  quizId: string
): Promise<number> {
  try {
    const supabase = await createClient();
    
    // Récupérer le dernier numéro d'attempt
    const { data: lastAttempt, error } = await supabase
      .from('user_quiz_attempts')
      .select('attempt_number')
      .eq('user_id', userId)
      .eq('quiz_id', quizId)
      .order('attempt_number', { ascending: false })
      .limit(1)
      .single();

    console.log('getUserAttemptNumber - lastAttempt:', lastAttempt, 'error:', error);
    
    if (lastAttempt) {
      return lastAttempt.attempt_number + 1;
    }
    return 1;
  } catch (error) {
    console.error('Erreur getUserAttemptNumber:', error);
    return 1;
  }
}

export async function recordQuizAttempt(
  userId: string,
  quizId: string,
  attemptNumber: number,
  score: number,
  xpEarned: number
): Promise<void> {
  try {
    const supabase = await createClient();
    
    const { error } = await supabase.from('user_quiz_attempts').insert({
      user_id: userId,
      quiz_id: quizId,
      attempt_number: attemptNumber,
      score,
      xp_earned: xpEarned,
    });

    if (error) {
      console.error('Erreur recordQuizAttempt:', error);
    } else {
      console.log('Quiz attempt enregistrée:', { userId, quizId, attemptNumber, score, xpEarned });
    }
  } catch (error) {
    console.error('Erreur recordQuizAttempt:', error);
  }
}

export async function recordQuestionAttempt(
  userId: string,
  quizId: string,
  questionId: string,
  attemptNumber: number,
  isCorrect: boolean,
  xpEarned: number
): Promise<void> {
  try {
    const supabase = await createClient();
    
    await supabase.from('user_question_attempts').insert({
      user_id: userId,
      quiz_id: quizId,
      question_id: questionId,
      attempt_number: attemptNumber,
      is_correct: isCorrect,
      xp_earned: xpEarned,
    });
  } catch (error) {
    console.error('Erreur recordQuestionAttempt:', error);
  }
}

export async function addXP(
  userId: string,
  amount: number,
  source: 'quiz' | 'streak' | 'challenge' | 'event',
  sourceId?: string
): Promise<void> {
  try {
    const supabase = await createClient();
    
    // Enregistrer la transaction
    await supabase.from('xp_transactions').insert({
      user_id: userId,
      source,
      source_id: sourceId || null,
      amount,
    });

    // Mettre à jour le total XP de l'utilisateur directement
    const { data: profile } = await supabase
      .from('user_profiles')
      .select('xp, total_xp')
      .eq('id', userId)
      .single();

    if (profile) {
      const newXP = (profile.xp || 0) + amount;
      const newTotalXP = (profile.total_xp || 0) + amount;
      const newLevel = Math.max(1, Math.floor(Math.sqrt(newXP / 10)) + 1);

      await supabase
        .from('user_profiles')
        .update({ 
          xp: newXP, 
          total_xp: newTotalXP,
          level: newLevel 
        })
        .eq('id', userId);

      // Mettre à jour le rang
      await updateUserRank(userId, newXP);
    }
  } catch (error) {
    console.error('Erreur addXP:', error);
  }
}

async function updateUserRank(userId: string, xp: number): Promise<void> {
  try {
    const supabase = await createClient();
    
    const { data: rankConfig } = await supabase
      .from('rank_config')
      .select('rank_label, xp_required')
      .order('display_order', { ascending: false });

    if (rankConfig) {
      const newRank = rankConfig.find(r => xp >= r.xp_required);
      if (newRank) {
        await supabase
          .from('user_profiles')
          .update({ rank: newRank.rank_label })
          .eq('id', userId);
      }
    }
  } catch (error) {
    console.error('Erreur updateUserRank:', error);
  }
}