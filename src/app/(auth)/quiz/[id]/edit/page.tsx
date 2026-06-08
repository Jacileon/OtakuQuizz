// ============================================================
// PAGE ÉDITION DE QUIZ
// ============================================================

import { redirect, notFound } from 'next/navigation';
import { createClient } from '@/lib/supabase/server';
import { getCurrentUser } from '@/lib/auth/actions';
import { QuizCreatorForm } from '@/components/quiz-creator/QuizCreatorForm';

export default async function QuizEditPage({
  params,
}: {
  params: { id: string };
}) {
  const user = await getCurrentUser();
  if (!user) redirect('/login');

  const supabase = createClient();

  // Charger le quiz
  const { data: quiz, error } = await supabase
    .from('quizzes')
    .select('*')
    .eq('id', params.id)
    .single();

  if (error || !quiz) {
    notFound();
  }

  if (quiz.creator_id !== user.id) {
    redirect(`/quiz/${params.id}`);
  }

  // Charger les questions séparément
  const { data: questions } = await supabase
    .from('questions')
    .select('*')
    .eq('quiz_id', params.id)
    .order('order_index', { ascending: true });

  // Charger les réponses pour chaque question
  const questionsWithAnswers = [];
  if (questions) {
    for (const q of questions) {
      const { data: answers } = await supabase
        .from('answers')
        .select('*')
        .eq('question_id', q.id)
        .order('order_index', { ascending: true });

      questionsWithAnswers.push({
        ...q,
        answers: answers || [],
      });
    }
  }

  const quizData = {
    ...quiz,
    questions: questionsWithAnswers,
  };

  console.log('Quiz data for edit:', JSON.stringify(quizData.questions?.map((q: any) => ({ id: q.id, text: q.question_text })), null, 2));

  return (
    <div className="p-4 lg:p-8 pb-24 md:pb-8">
      <div className="max-w-4xl mx-auto">
        <h1 className="font-display text-3xl tracking-wider mb-2">MODIFIER LE QUIZ</h1>
        <p className="text-muted-foreground mb-8">
          Modifie les détails de ton quiz
        </p>
        <QuizCreatorForm quizId={params.id} initialData={quizData as any} />
      </div>
    </div>
  );
}