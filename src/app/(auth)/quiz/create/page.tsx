// ============================================================
// PAGE CRÉATION DE QUIZ
// ============================================================

import { redirect } from '../../../../../node_modules/next/navigation';
import { getCurrentUser } from '@/lib/auth/actions';
import { canUserCreateQuiz } from '@/lib/actions/permissions';
import { QuizCreatorForm } from '@/components/quiz-creator/QuizCreatorForm';
import { Card, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Lock, ArrowLeft } from 'lucide-react';
import Link from 'next/link';

export default async function QuizCreatePage() {
  const user = await getCurrentUser();
  if (!user) redirect('/login');

  const { allowed, reason } = await canUserCreateQuiz();

  if (!allowed) {
    return (
      <div className="p-4 lg:p-8 pb-24 md:pb-8">
        <div className="max-w-md mx-auto">
          <Card className="border-destructive/30">
            <CardContent className="p-8 text-center">
              <Lock className="h-12 w-12 text-destructive mx-auto mb-4" />
              <h2 className="text-xl font-semibold mb-2">Accès restreint</h2>
              <p className="text-muted-foreground mb-6">{reason}</p>
              <Link href="/dashboard">
                <Button variant="outline" className="gap-2">
                  <ArrowLeft className="h-4 w-4" />
                  Retour au tableau de bord
                </Button>
              </Link>
            </CardContent>
          </Card>
        </div>
      </div>
    );
  }

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