'use client';

import { Quiz, PaginatedResponse } from '@/types';
import { QuizCard } from '@/components/dashboard/QuizCard';
import { Button } from '@/components/ui/button';
import { ChevronLeft, ChevronRight, Search } from 'lucide-react';
import Link from '../../../node_modules/next/link';
import { useSearchParams } from '../../../node_modules/next/navigation';

interface QuizGridProps {
  quizzes: Quiz[];
  pagination: PaginatedResponse<Quiz>;
}

export function QuizGrid({ quizzes, pagination }: QuizGridProps) {
  const searchParams = useSearchParams();

  const createPageUrl = (page: number) => {
    const params = new URLSearchParams(searchParams);
    params.set('page', page.toString());
    return `?${params.toString()}`;
  };

  if (quizzes.length === 0) {
    return (
      <div className="text-center py-20">
        <Search className="h-12 w-12 mx-auto mb-4 text-muted-foreground" />
        <h3 className="font-display text-xl mb-2">Aucun quiz trouvé</h3>
        <p className="text-muted-foreground">Essaye avec d'autres critères de recherche</p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="grid sm:grid-cols-2 lg:grid-cols-3 gap-6">
        {quizzes.map((quiz) => (
          <QuizCard key={quiz.id} quiz={quiz} />
        ))}
      </div>

      {/* Pagination */}
      {pagination.total_pages > 1 && (
        <div className="flex items-center justify-center gap-2 pt-6">
          <Link href={createPageUrl(pagination.page - 1)}>
            <Button variant="outline" size="icon" disabled={pagination.page <= 1}>
              <ChevronLeft className="h-4 w-4" />
            </Button>
          </Link>

          <span className="text-sm text-muted-foreground px-4">
            Page {pagination.page} / {pagination.total_pages}
          </span>

          <Link href={createPageUrl(pagination.page + 1)}>
            <Button variant="outline" size="icon" disabled={pagination.page >= pagination.total_pages}>
              <ChevronRight className="h-4 w-4" />
            </Button>
          </Link>
        </div>
      )}
    </div>
  );
}

