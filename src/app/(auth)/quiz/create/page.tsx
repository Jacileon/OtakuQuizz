// ============================================================
// PAGE CRÉATION DE QUIZ
// ============================================================

import { redirect } from '../../../../../node_modules/next/navigation';
import { getCurrentUser } from '@/lib/auth/actions';
import { QuizCreatorForm } from '@/components/quiz-creator/QuizCreatorForm';

export default async function QuizCreatePage() {
  const user = await getCurrentUser();
  if (!user) redirect('/login');

  return (
    <div className="p-4 lg:p-8 pb-24 md:pb-8">
      <div className="max-w-4xl mx-auto">
        <h1 className="font-display text-3xl tracking-wider mb-2">CRÉER UN QUIZ</h1>
        <p className="text-muted-foreground mb-8">
          Partage ta passion avec la communauté otaku africaine
        </p>
        <QuizCreatorForm />
      </div>
    </div>
  );
}

