'use client';

import { QuestionClient, GuessData } from '@/types';
import { cn } from '@/lib/utils';
import { Headphones, Image, FileText, HelpCircle, Users, Search } from 'lucide-react';

interface QuestionDisplayProps {
  question: QuestionClient;
}

export function QuestionDisplay({ question }: QuestionDisplayProps) {
  const renderMedia = () => {
    switch (question.question_type) {
      case 'image':
  return (
    <div className="relative rounded-lg overflow-hidden mb-4">
      <img
        src={question.media_url || ''}
        className="w-full max-h-64 object-contain rounded-lg"
      />
    </div>
  );

      case 'gif':
        return (
          <div className="rounded-lg overflow-hidden mb-4">
            <img
              src={question.media_url || ''}
              alt="GIF"
              className="w-full max-h-64 object-contain rounded-lg"
            />
          </div>
        );

      case 'audio':
        return (
          <div className="mb-4 p-4 rounded-lg bg-dark-surface border border-dark-border">
            <div className="flex items-center gap-3">
              <Headphones className="h-8 w-8 text-brand" />
              <div>
                <p className="text-sm font-medium">Extrait audio</p>
                <p className="text-xs text-muted-foreground">Écoute attentivement...</p>
              </div>
            </div>
            {question.media_url && (
              <audio
                src={question.media_url}
                controls
                className="w-full mt-3"
                autoPlay
              />
            )}
          </div>
        );

      case 'true_false':
        return (
          <div className="flex items-center gap-2 mb-4 text-accent">
            <HelpCircle className="h-5 w-5" />
            <span className="text-sm font-medium">Vrai ou Faux ?</span>
          </div>
        );

      case 'character_guess':
        const characters = question.character_guess_data?.characters || [];
        const mode = question.character_guess_mode || 'image';
        return (
          <div className="mb-4">
            <div className="flex items-center gap-2 mb-3">
              <Users className="h-5 w-5 text-brand" />
              <span className="text-sm font-medium text-muted-foreground">Devine le personnage/anime</span>
            </div>
            <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
              {characters.slice(0, 4).map((char, i) => (
                <div key={i} className="p-3 rounded-lg bg-dark-surface border border-dark-border">
                  <span className="text-xs font-medium text-brand mb-2 block">Indice {i + 1}</span>
                  {mode === 'image' && char.image_url ? (
                    <img
                      src={char.image_url}
                      alt={`Indice ${i + 1}`}
                      className="w-full h-24 object-cover rounded-lg"
                    />
                  ) : (
                    <p className="text-sm text-white">{char.clue || char.image_url || '?'}</p>
                  )}
                </div>
              ))}
            </div>
          </div>
        );

      case 'impostor':
        const oddItems = question.find_odd_data?.items || [];
        return (
          <div className="mb-4">
            <div className="flex items-center gap-2 mb-3">
              <Search className="h-5 w-5 text-brand" />
              <span className="text-sm font-medium text-muted-foreground">Trouve l'intrus parmi ces {oddItems.length} éléments :</span>
            </div>
          </div>
        );

      default:
        return null;
    }
  };

  return (
    <div className="space-y-4">
      {renderMedia()}
      <div className="flex items-start gap-3">
        <FileText className="h-5 w-5 text-brand mt-1 shrink-0" />
        <p className="text-lg font-medium leading-relaxed">{question.question_text}</p>
      </div>
    </div>
  );
}

