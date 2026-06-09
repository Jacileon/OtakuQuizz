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
    return Math.round(baseXP / 2 * 100) / 100;
  } else if (attemptNumber === 3) {
    return Math.round(baseXP / 4 * 100) / 100;
  } else {
    return 0;
  }
}

export async function getUserAttemptNumber(
  userId: string,
  quizId: string
): Promise<number> {
  const supabase = await createClient();
  const { count } = await supabase
    .from('user_quiz_attempts')
    .select('*', { count: 'exact', head: true })
    .eq('user_id', userId)
    .eq('quiz_id', quizId);

  return (count || 0) + 1;
}

export async function recordQuizAttempt(
  userId: string,
  quizId: string,
  attemptNumber: number,
  score: number,
  xpEarned: number
): Promise<void> {
  const supabase = await createClient();
  
  await supabase.from('user_quiz_attempts').insert({
    user_id: userId,
    quiz_id: quizId,
    attempt_number: attemptNumber,
    score,
    xp_earned: xpEarned,
  });
}

export async function recordQuestionAttempt(
  userId: string,
  quizId: string,
  questionId: string,
  attemptNumber: number,
  isCorrect: boolean,
  xpEarned: number
): Promise<void> {
  const supabase = await createClient();
  
  await supabase.from('user_question_attempts').insert({
    user_id: userId,
    quiz_id: quizId,
    question_id: questionId,
    attempt_number: attemptNumber,
    is_correct: isCorrect,
    xp_earned: xpEarned,
  });
}

export async function addXP(
  userId: string,
  amount: number,
  source: 'quiz' | 'streak' | 'challenge' | 'event',
  sourceId?: string
): Promise<void> {
  const supabase = await createClient();
  
  // Enregistrer la transaction
  await supabase.from('xp_transactions').insert({
    user_id: userId,
    source,
    source_id: sourceId || null,
    amount,
  });

  // Mettre à jour le total XP de l'utilisateur
  await supabase.rpc('increment_user_xp', {
    user_id: userId,
    amount: amount,
  });
}