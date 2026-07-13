'use client';

import { useState } from "react";
import { useRouter, useSearchParams } from '../../../node_modules/next/navigation';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import { Search, SlidersHorizontal, X } from 'lucide-react';
import { SUBCATEGORY_LIST } from '@/lib/constants';
import { cn } from '@/lib/utils';

interface FilterBarProps {
  categories: string[];
  series: string[];
  currentParams: Record<string, string | undefined>;
}

export function FilterBar({ categories, series, currentParams }: FilterBarProps) {
  const router = useRouter();
  const searchParams = useSearchParams();
  const [showFilters, setShowFilters] = useState(false);
  const [searchQuery, setSearchQuery] = useState(currentParams.q || '');

  const currentCategory = currentParams.category || '';
  const currentSubcategory = currentParams.subcategory || '';
  const currentSort = currentParams.sort || 'popular';

  const updateSearch = (updates: Record<string, string | undefined>) => {
    const params = new URLSearchParams(searchParams);
    Object.entries(updates).forEach(([key, value]) => {
      if (value) params.set(key, value);
      else params.delete(key);
    });
    params.set('page', '1');
    router.push(`?${params.toString()}`);
  };

  const handleSearch = () => {
    updateSearch({ q: searchQuery || undefined });
  };

  const sortOptions = [
    { value: 'popular', label: 'Plus joués' },
    { value: 'recent', label: 'Plus récents' },
    { value: 'rated', label: 'Mieux notés' },
  ];

  return (
    <div className="space-y-4">
      {/* Search bar */}
      <div className="flex gap-2">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            placeholder="Rechercher un quiz, une série..."
            className="pl-10 bg-dark-surface border-dark-border"
            onKeyDown={(e) => e.key === 'Enter' && handleSearch()}
          />
        </div>
        <Button onClick={handleSearch}>
          <Search className="h-4 w-4" />
        </Button>
        <Button variant="outline" onClick={() => setShowFilters(!showFilters)}>
          <SlidersHorizontal className="h-4 w-4" />
        </Button>
      </div>

      {/* Filters */}
      {showFilters && (
        <div className="p-4 rounded-lg bg-dark-card border border-dark-border space-y-4">
          {/* Categories */}
          <div>
            <span className="text-sm font-medium mb-2 block">Catégorie</span>
            <div className="flex flex-wrap gap-2">
              <button
                onClick={() => updateSearch({ category: undefined, subcategory: undefined })}
                className={cn(
                  'px-3 py-1.5 rounded-md text-sm transition-colors',
                  !currentCategory ? 'bg-brand text-white' : 'bg-dark-surface text-muted-foreground hover:text-white'
                )}
              >
                Toutes
              </button>
              {categories.map((cat) => (
                <button
                  key={cat}
                  onClick={() => updateSearch({ category: cat, subcategory: undefined })}
                  className={cn(
                    'px-3 py-1.5 rounded-md text-sm transition-colors',
                    currentCategory === cat ? 'bg-brand text-white' : 'bg-dark-surface text-muted-foreground hover:text-white'
                  )}
                >
                  {cat}
                </button>
              ))}
            </div>
          </div>

          {/* Subcategories */}
          {currentCategory && SUBCATEGORY_LIST[currentCategory] && (
            <div>
              <span className="text-sm font-medium mb-2 block">Sous-catégorie</span>
              <div className="flex flex-wrap gap-2">
                <button
                  onClick={() => updateSearch({ subcategory: undefined })}
                  className={cn(
                    'px-3 py-1.5 rounded-md text-sm transition-colors',
                    !currentSubcategory ? 'bg-brand text-white' : 'bg-dark-surface text-muted-foreground hover:text-white'
                  )}
                >
                  Toutes
                </button>
                {SUBCATEGORY_LIST[currentCategory].map((sub) => (
                  <button
                    key={sub}
                    onClick={() => updateSearch({ subcategory: sub })}
                    className={cn(
                      'px-3 py-1.5 rounded-md text-sm transition-colors',
                      currentSubcategory === sub ? 'bg-brand text-white' : 'bg-dark-surface text-muted-foreground hover:text-white'
                    )}
                  >
                    {sub}
                  </button>
                ))}
              </div>
            </div>
          )}

          {/* Sort */}
          <div>
            <span className="text-sm font-medium mb-2 block">Trier par</span>
            <div className="flex flex-wrap gap-2">
              {sortOptions.map((opt) => (
                <button
                  key={opt.value}
                  onClick={() => updateSearch({ sort: opt.value })}
                  className={cn(
                    'px-3 py-1.5 rounded-md text-sm transition-colors',
                    currentSort === opt.value ? 'bg-brand text-white' : 'bg-dark-surface text-muted-foreground hover:text-white'
                  )}
                >
                  {opt.label}
                </button>
              ))}
            </div>
          </div>
        </div>
      )}

      {/* Active filters */}
      {(currentCategory || currentParams.q) && (
        <div className="flex flex-wrap gap-2">
          {currentCategory && (
            <Badge variant="secondary" className="gap-1 cursor-pointer" onClick={() => updateSearch({ category: undefined })}>
              {currentCategory} <X className="h-3 w-3" />
            </Badge>
          )}
          {currentParams.q && (
            <Badge variant="secondary" className="gap-1 cursor-pointer" onClick={() => updateSearch({ q: undefined })}>
              "{currentParams.q}" <X className="h-3 w-3" />
            </Badge>
          )}
        </div>
      )}
    </div>
  );
}

