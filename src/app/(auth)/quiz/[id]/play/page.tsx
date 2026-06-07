// ============================================================
// PAGE DE JEU - Server Component
// ============================================================

import { redirect } from '../../../../../../node_modules/next/navigation';
import { createClient } from '@/lib/supabase/server';
import { getCurrentUser } from '@/lib/auth/actions';
import { QuizEngine } from '@/components/quiz/QuizEngine';
import { QuestionClient } from '@/types';
import { shuffleArray } from '@/lib/utils';

export default async function QuizPlayPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const user = await getCurrentUser();
  if (!user) redirect('/login');

  const supabase = createClient();

  // 1. Vérifier que le quiz existe et est accessible
  const { data: quiz } = await supabase
    .from('quizzes')
    .select('*')
    .eq('id', id)
    .single();

  if (!quiz || !quiz.is_visible) {
    redirect('/explore');
  }

  // 2. Pour les quiz officiels, vérifier la fenêtre temporelle
  if (quiz.quiz_type === 'official') {
    const now = new Date();
    const start = quiz.event_start_at ? new Date(quiz.event_start_at) : null;
    const end = quiz.event_end_at ? new Date(quiz.event_end_at) : null;

    if (start && now < start) {
      redirect(`/events/${id}`); // Pas encore commencé
    }
    if (end && now > end) {
      redirect(`/events/${id}`); // Terminé
    }
  }

  // 3. Créer une game_session
  const { data: session } = await supabase
    .from('game_sessions')
    .insert({
      user_id: user.id,
      quiz_id: id,
      total_questions: quiz.question_count,
    })
    .select()
    .single();

  if (!session) {
    redirect('/explore');
  }

  // 4. Charger les questions avec réponses SANS is_correct
  const { data: questions } = await supabase
    .from('questions')
    .select(`
      id, quiz_id, question_text, question_type, media_url, media_public_id,
      time_limit_seconds, order_index, character_guess_data, character_guess_mode,
      find_odd_data,
      answers:answers(id, question_id, answer_text, order_index)
    `)
    .eq('quiz_id', id)
    .order('order_index', { ascending: true });

  if (!questions || questions.length === 0) {
    redirect('/explore');
  }

  // 5. Randomiser l'ordre des questions et réponses côté serveur
  const randomizedQuestions: QuestionClient[] = shuffleArray(questions).map((q: any) => ({
    ...q,
    answers: shuffleArray(q.answers.map((a: any) => ({
      id: a.id,
      question_id: a.question_id,
      answer_text: a.answer_text,
      order_index: a.order_index,
    }))),
  }));

  return (
    <QuizEngine
      quizId={id}
      sessionId={session.id}
      questions={randomizedQuestions}
      isOfficial={quiz.quiz_type === 'official'}
      quizTitle={quiz.title}
    />
  );
}
