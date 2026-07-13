// ============================================================
// PAGE MODÉRATION ADMIN
// ============================================================

import { createClient } from '@/lib/supabase/server';
import { requireAdmin } from '@/lib/auth/actions';
import { Card, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Shield, AlertTriangle, Eye, Trash2, RotateCcw } from 'lucide-react';
import { redirect } from '../../../../../node_modules/next/navigation';

export default async function AdminReportsPage() {
  await requireAdmin();

  const supabase = createClient();

  const { data: reportedQuizzes } = await supabase
    .from('quizzes')
    .select('*, reports:reports(count), creator:creator_id(username)')
    .eq('is_visible', false)
    .gt('total_reports', 0)
    .order('total_reports', { ascending: false });

  const { data: pendingReports } = await supabase
    .from('reports')
    .select('*, quiz:quiz_id(title), reporter:reporter_id(username)')
    .eq('status', 'pending')
    .order('created_at', { ascending: false });

  return (
    <div className="p-4 lg:p-8">
      <div className="max-w-6xl mx-auto space-y-8">
        <h1 className="font-display text-3xl tracking-wider flex items-center gap-3">
          <Shield className="h-8 w-8 text-brand" />
          MODÉRATION
        </h1>

        <Tabs defaultValue="hidden" className="w-full">
          <TabsList className="w-full justify-start">
            <TabsTrigger value="hidden" className="gap-2">
              <AlertTriangle className="h-4 w-4" /> Quiz masqués ({reportedQuizzes?.length || 0})
            </TabsTrigger>
            <TabsTrigger value="reports" className="gap-2">
              <Shield className="h-4 w-4" /> Signalements ({pendingReports?.length || 0})
            </TabsTrigger>
          </TabsList>

          <TabsContent value="hidden" className="space-y-4">
            {reportedQuizzes?.map((quiz: any) => (
              <Card key={quiz.id} className="border-red-500/30 bg-red-500/5">
                <CardContent className="p-4">
                  <div className="flex items-start justify-between">
                    <div>
                      <div className="flex items-center gap-2">
                        <h3 className="font-medium">{quiz.title}</h3>
                        <Badge variant="destructive">{quiz.total_reports} signalements</Badge>
                      </div>
                      <p className="text-sm text-muted-foreground mt-1">
                        Par {quiz.creator?.username} • {quiz.series}
                      </p>
                    </div>
                    <div className="flex gap-2">
                      <Button size="sm" variant="outline" className="gap-1">
                        <Eye className="h-3.5 w-3.5" /> Voir
                      </Button>
                      <Button size="sm" variant="outline" className="gap-1 text-green-400">
                        <RotateCcw className="h-3.5 w-3.5" /> Restaurer
                      </Button>
                      <Button size="sm" variant="destructive" className="gap-1">
                        <Trash2 className="h-3.5 w-3.5" /> Supprimer
                      </Button>
                    </div>
                  </div>
                </CardContent>
              </Card>
            )) || <p className="text-muted-foreground">Aucun quiz masqué</p>}
          </TabsContent>

          <TabsContent value="reports" className="space-y-4">
            {pendingReports?.map((report: any) => (
              <Card key={report.id} className="border-dark-border bg-dark-card">
                <CardContent className="p-4">
                  <div className="flex items-start justify-between">
                    <div>
                      <div className="flex items-center gap-2">
                        <Badge>{report.reason}</Badge>
                        <span className="text-sm text-muted-foreground">
                          Sur "{report.quiz?.title}"
                        </span>
                      </div>
                      <p className="text-sm mt-1">{report.description}</p>
                      <p className="text-xs text-muted-foreground mt-1">
                        Par {report.reporter?.username}
                      </p>
                    </div>
                    <div className="flex gap-2">
                      <Button size="sm">Résoudre</Button>
                      <Button size="sm" variant="ghost">Ignorer</Button>
                    </div>
                  </div>
                </CardContent>
              </Card>
            )) || <p className="text-muted-foreground">Aucun signalement en attente</p>}
          </TabsContent>
        </Tabs>
      </div>
    </div>
  );
}

