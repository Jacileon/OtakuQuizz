import Link from '../../../node_modules/next/link';
import { Quiz } from '@/types';
import { Card, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Gamepad2, HelpCircle, Play, Eye } from 'lucide-react';

interface QuizCardProps {
  quiz: Quiz;
}

export function QuizCard({ quiz }: QuizCardProps) {
  return (
    <Card className="border-dark-border bg-dark-card/50 hover:border-brand/30 transition-all overflow-hidden group">
      <div className="h-32 bg-gradient-to-br from-dark-surface to-dark-card flex items-center justify-center relative">
        {quiz.thumbnail_url ? (
          <img src={quiz.thumbnail_url} alt={quiz.title} className="h-full w-full object-cover" />
        ) : (
          <Gamepad2 className="h-12 w-12 text-white/10 group-hover:text-white/20 transition-colors" />
        )}
        <div className="absolute top-2 left-2">
          <Badge variant="secondary" className="text-xs">
            {quiz.series}
          </Badge>
        </div>
      </div>
      <CardContent className="p-4">
        <h3 className="font-medium mb-1 truncate">{quiz.title}</h3>
        <div className="flex items-center gap-3 text-xs text-muted-foreground mb-3">
          <span className="flex items-center gap-1">
            <HelpCircle className="h-3 w-3" />
            {quiz.question_count} questions
          </span>
          <span>{quiz.play_count} plays</span>
        </div>
        <div className="flex gap-2">
          <Link href={`/quiz/${quiz.id}/play`} className="flex-1">
            <Button className="w-full gap-2" size="sm">
              <Play className="h-4 w-4" />
              Jouer
            </Button>
          </Link>
          <Link href={`/quiz/${quiz.id}`}>
            <Button variant="outline" size="sm" className="gap-1">
              <Eye className="h-4 w-4" />
              Voir
            </Button>
          </Link>
        </div>
      </CardContent>
    </Card>
  );
}