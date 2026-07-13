// ============================================================
// PAGE CATÉGORIE
// ============================================================

import { searchQuizzes } from '@/lib/queries/quizzes';
import { QuizGrid } from '@/components/explore/QuizGrid';
import { FilterBar } from '@/components/explore/FilterBar';
import { CATEGORY_LIST, SUBCATEGORY_LIST } from '@/lib/constants';
import { Card, CardContent } from '@/components/ui/card';
import { Gamepad2 } from 'lucide-react';

export default async function CategoryPage({ params }: { params: { category: string } }) {
  const result = await searchQuizzes({ category: params.category });
  const subcategories = SUBCATEGORY_LIST[params.category] || [];

  return (
    <div className="p-4 lg:p-8 pb-24 md:pb-8">
      <div className="max-w-6xl mx-auto space-y-6">
        <div className="relative overflow-hidden rounded-lg bg-gradient-to-r from-brand/20 to-accent/20 p-8">
          <Gamepad2 className="h-16 w-16 text-white/10 absolute right-4 top-4" />
          <h1 className="font-display text-4xl tracking-wider">{params.category}</h1>
          <p className="text-muted-foreground mt-2">{result.count} quiz disponibles</p>
        </div>

        <div className="flex flex-wrap gap-2">
          {subcategories.map((sub) => (
            <a
              key={sub}
              href={`/explore?category=${params.category}&subcategory=${sub}`}
              className="px-3 py-1.5 rounded-md bg-dark-surface border border-dark-border text-sm hover:border-brand/50 transition-colors"
            >
              {sub}
            </a>
          ))}
        </div>

        <QuizGrid quizzes={result.data} pagination={result} />
      </div>
    </div>
  );
}
