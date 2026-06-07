// ============================================================
// ROUTE API - SOUMISSION DES RÉPONSES (CRITIQUE)
// Scores calculés UNIQUEMENT côté serveur
// ============================================================

import { NextResponse } from '../../../../../node_modules/next/server';
import { createClient } from '@/lib/supabase/server';
import { calculateScore } from '@/lib/utils';
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

    if (!sessionId || !Array.isArray(answers)) {
      return NextResponse.json({ error: 'Données invalides' }, { status: 400 });
    }

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

    // Récupérer les bonnes réponses depuis la BDD (jamais exposées au client)
    const questionIds = answers.map((a: any) => a.questionId);
    const { data: correctAnswers } = await supabase
      .from('answers')
      .select('id, question_id, is_correct')
      .in('question_id', questionIds);

    const correctMap = new Map();
    correctAnswers?.forEach((a) => {
      if (a.is_correct) correctMap.set(a.question_id, a.id);
    });

    // Calculer le score pour chaque réponse
    let totalScore = 0;
    let correctCount = 0;
    let streak = 0;
    let maxStreak = 0;
    const playerAnswers = [];

    for (const answer of answers) {
      const correctAnswerId = correctMap.get(answer.questionId);
      const isCorrect = answer.answerId === correctAnswerId;
      const question = await supabase
        .from('questions')
        .select('time_limit_seconds')
        .eq('id', answer.questionId)
        .single();

      const timeLimitMs = (question.data?.time_limit_seconds || 30) * 1000;

      if (isCorrect) {
        streak++;
        maxStreak = Math.max(maxStreak, streak);
        correctCount++;
      } else {
        streak = 0;
      }

      const points = calculateScore(isCorrect, answer.timeMs, timeLimitMs, streak);
      totalScore += points;

      playerAnswers.push({
        session_id: sessionId,
        question_id: answer.questionId,
        answer_id: answer.answerId,
        is_correct: isCorrect,
        time_taken_ms: answer.timeMs,
        points_earned: points,
      });
    }

    const totalQuestions = answers.length;
    const accuracyRate = totalQuestions > 0 ? Math.round((correctCount / totalQuestions) * 100) : 0;
    const isPerfect = correctCount === totalQuestions;
    const totalTimeMs = answers.reduce((sum: number, a: any) => sum + a.timeMs, 0);

    // Enregistrer les réponses en base
    await supabase.from('player_answers').insert(playerAnswers);

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

    // Vérifier et attribuer les badges (via fonction SQL)
    const { data: newBadges } = await supabase
      .rpc('check_and_award_badges', { target_user_id: user.id });

    const result: QuizSubmitResult = {
      score: totalScore,
      correctCount,
      totalQuestions,
      accuracyRate: accuracyRate,
      isPerfect,
      xpEarned: totalScore,
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

