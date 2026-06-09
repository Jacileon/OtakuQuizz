// ============================================================
// ROUTE API - SOUMISSION DES RÉPONSES (CRITIQUE)
// Scores calculés UNIQUEMENT côté serveur
// ============================================================

import { NextResponse } from '../../../../../node_modules/next/server';
import { createClient } from '@/lib/supabase/server';
import { calculateScore } from '@/lib/utils';
import { calculateQuestionXP, getUserAttemptNumber, recordQuizAttempt, recordQuestionAttempt, addXP } from '@/lib/utils/xp';
import { QuizSubmitResult } from '@/types';

export async function POST(request: Request) {
  try {
    const supabase = createClient();
    const { data: { user } } = await supabase.auth.getUser();

    if (!user) {
      return NextResponse.json({ error: 'Non authentifié' }, { status: 401 });
    }

    const body = await request.json();
    const { sessionId, answers } = body;

    if (!sessionId) {
      return NextResponse.json({ error: 'Données invalides' }, { status: 400 });
    }

    // Permettre un tableau vide (aucune réponse)
    const safeAnswers = Array.isArray(answers) ? answers : [];

    // Vérifier que la session appartient à l'utilisateur
    const { data: session } = await supabase
      .from('game_sessions')
      .select('*')
      .eq('id', sessionId)
      .eq('user_id', user.id)
      .single();

    if (!session) {
      return NextResponse.json({ error: 'Session non trouvée' }, { status: 404 });
    }

    if (session.completed_at) {
      return NextResponse.json({ error: 'Session déjà complétée' }, { status: 400 });
    }

    // Vérifier si une tentative est déjà en cours (éviter les doubles soumissions)
    const { data: existingAttempt } = await supabase
      .from('user_quiz_attempts')
      .select('id')
      .eq('user_id', user.id)
      .eq('quiz_id', session.quiz_id)
      .eq('attempt_number', await getUserAttemptNumber(user.id, session.quiz_id))
      .single();

    if (existingAttempt) {
      return NextResponse.json({ error: 'Tentative déjà enregistrée' }, { status: 400 });
    }

    // Récupérer les bonnes réponses depuis la BDD (jamais exposées au client)
    const questionIds = safeAnswers.map((a: any) => a.questionId);
    const { data: correctAnswers } = questionIds.length > 0 ? await supabase
      .from('answers')
      .select('id, question_id, is_correct')
      .in('question_id', questionIds) : { data: [] };

    const correctMap = new Map();
    correctAnswers?.forEach((a) => {
      if (a.is_correct) correctMap.set(a.question_id, a.id);
    });

    // Récupérer les types de questions
    const { data: questionsData } = questionIds.length > 0 ? await supabase
      .from('questions')
      .select('id, question_type, time_limit_seconds')
      .in('id', questionIds) : { data: [] };

    const questionTypeMap = new Map();
    questionsData?.forEach((q) => {
      questionTypeMap.set(q.id, { type: q.question_type, timeLimit: q.time_limit_seconds || 30 });
    });

    // Numéro de tentative
    const attemptNumber = await getUserAttemptNumber(user.id, session.quiz_id);

    // Calculer le score et l'XP pour chaque réponse
    let totalScore = 0;
    let totalXP = 0;
    let correctCount = 0;
    let streak = 0;
    let maxStreak = 0;
    const playerAnswers = [];

    for (const answer of safeAnswers) {
      const correctAnswerId = correctMap.get(answer.questionId);
      const isCorrect = answer.answerId === correctAnswerId;
      const questionInfo = questionTypeMap.get(answer.questionId);
      const questionType = questionInfo?.type || 'text';
      const timeLimitMs = (questionInfo?.timeLimit || 30) * 1000;

      if (isCorrect) {
        streak++;
        maxStreak = Math.max(maxStreak, streak);
        correctCount++;
      } else {
        streak = 0;
      }

      const points = calculateScore(isCorrect, answer.timeMs, timeLimitMs, streak);
      totalScore += points;

      // Calculer l'XP pour cette question
      const xpForQuestion = await calculateQuestionXP(
        user.id,
        session.quiz_id,
        answer.questionId,
        questionType,
        isCorrect,
        attemptNumber
      );
      totalXP += Math.floor(xpForQuestion);

      // Enregistrer la tentative de question
      await recordQuestionAttempt(
        user.id,
        session.quiz_id,
        answer.questionId,
        attemptNumber,
        isCorrect,
        xpForQuestion
      );

      playerAnswers.push({
        session_id: sessionId,
        question_id: answer.questionId,
        answer_id: answer.answerId,
        is_correct: isCorrect,
        time_taken_ms: answer.timeMs,
        points_earned: points,
      });
    }

    const totalQuestions = session.total_questions;
    const accuracyRate = totalQuestions > 0 ? Math.round((correctCount / totalQuestions) * 100) : 0;
    const isPerfect = correctCount === totalQuestions && totalQuestions > 0;
    const totalTimeMs = safeAnswers.reduce((sum: number, a: any) => sum + (a.timeMs || 0), 0);

    // Enregistrer les réponses en base (seulement s'il y en a)
    if (playerAnswers.length > 0) {
      await supabase.from('player_answers').insert(playerAnswers);
    }

    // Mettre à jour la session
    await supabase
      .from('game_sessions')
      .update({
        completed_at: new Date().toISOString(),
        score: totalScore,
        correct_count: correctCount,
        accuracy_rate: accuracyRate,
        is_perfect: isPerfect,
        time_taken_ms: totalTimeMs,
      })
      .eq('id', sessionId);

    // Enregistrer la tentative de quiz
    const finalXP = Math.floor(totalXP);
    console.log('Recording quiz attempt:', { userId: user.id, quizId: session.quiz_id, attemptNumber, totalScore, finalXP });
    await recordQuizAttempt(user.id, session.quiz_id, attemptNumber, totalScore, finalXP);

    // Attribuer l'XP
    console.log('Adding XP:', { userId: user.id, finalXP });
    if (finalXP > 0) {
      await addXP(user.id, finalXP, 'quiz', session.quiz_id);
    }

    // Mettre à jour le classement du quiz
    console.log('Updating leaderboard');
    const { error: leaderboardError } = await supabase.from('leaderboard_monthly').upsert({
      user_id: user.id,
      month_year: new Date().toISOString().slice(0, 7),
      score: totalScore,
    }, { onConflict: 'user_id,month_year' });

    if (leaderboardError) {
      console.error('Erreur leaderboard:', leaderboardError);
    } else {
      console.log('Leaderboard mis à jour avec succès');
    }

    // Mettre à jour les stats utilisateur
    console.log('Updating user stats');
    const { data: existingStats } = await supabase
      .from('user_stats')
      .select('*')
      .eq('user_id', user.id)
      .single();

    if (existingStats) {
      // Mettre à jour les stats existantes
      await supabase
        .from('user_stats')
        .update({
          quizzes_played: (existingStats.quizzes_played || 0) + 1,
          total_correct_answers: (existingStats.total_correct_answers || 0) + correctCount,
          total_answers: (existingStats.total_answers || 0) + totalQuestions,
          accuracy_rate: Math.round(((existingStats.total_correct_answers || 0) + correctCount) / ((existingStats.total_answers || 0) + totalQuestions) * 100),
          best_score_ever: Math.max(existingStats.best_score_ever || 0, totalScore),
          updated_at: new Date().toISOString(),
        })
        .eq('user_id', user.id);
    } else {
      // Créer les stats
      await supabase
        .from('user_stats')
        .insert({
          user_id: user.id,
          quizzes_played: 1,
          total_correct_answers: correctCount,
          total_answers: totalQuestions,
          accuracy_rate: accuracyRate,
          best_score_ever: totalScore,
        });
    }
    console.log('User stats updated');

    // Vérifier et attribuer les badges (via fonction SQL)
    const { data: newBadges } = await supabase
      .rpc('check_and_award_badges', { target_user_id: user.id });

    const result: QuizSubmitResult = {
      score: totalScore,
      correctCount,
      totalQuestions,
      accuracyRate: accuracyRate,
      isPerfect,
      xpEarned: totalXP,
      newBadges: newBadges || [],
    };

    return NextResponse.json(result);
  } catch (error) {
    console.error('Erreur submit quiz:', error);
    return NextResponse.json(
      { error: 'Erreur serveur' },
      { status: 500 }
    );
  }
}