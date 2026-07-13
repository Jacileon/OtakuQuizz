// ============================================================
// PAGE EXPLORATION
// ============================================================

import { Suspense } from 'react';
import { searchQuizzes, getAllSeries } from '@/lib/queries/quizzes';
import { QuizGrid } from '@/components/explore/QuizGrid';
import { FilterBar } from '@/components/explore/FilterBar';
import { Skeleton } from '@/components/ui/skeleton';
import { CATEGORY_LIST } from '@/lib/constants';

interface ExplorePageProps {
  searchParams: {
    q?: string;
    category?: string;
    subcategory?: string;
    series?: string;
    sort?: string;
    page?: string;
  };
}

export default async function ExplorePage({ searchParams }: ExplorePageProps) {
  const params = {
    query: searchParams.q,
    category: searchParams.category,
    subcategory: searchParams.subcategory,
    series: searchParams.series,
    sortBy: (searchParams.sort as any) || 'popular',
    page: parseInt(searchParams.page || '1'),
  };

  const [result, series] = await Promise.all([
    searchQuizzes(params),
    getAllSeries(),
  ]);

  return (
    <div className="p-4 lg:p-8 pb-24 md:pb-8">
      <div className="max-w-6xl mx-auto space-y-6">
        <h1 className="font-display text-3xl tracking-wider">EXPLORER</h1>

        <FilterBar 
          categories={CATEGORY_LIST} 
          series={series} 
          currentParams={searchParams} 
        />

        <Suspense fallback={
          <div className="grid sm:grid-cols-2 lg:grid-cols-3 gap-6">
            {Array.from({ length: 6 }).map((_, i) => (
              <Skeleton key={i} className="h-64 rounded-lg" />
            ))}
          </div>
        }>
          <QuizGrid quizzes={result.data} pagination={result} />
        </Suspense>
      </div>
    </div>
  );
}

