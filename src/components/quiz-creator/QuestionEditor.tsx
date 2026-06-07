'use client';


import { QuestionCreateInput, CharacterGuessItem, GuessClue, FindOddItem } from '@/types';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';

import { Trash2, GripVertical, Image, Headphones, FileText, HelpCircle, Users, Search } from 'lucide-react';
import { cn } from '@/lib/utils';

interface QuestionEditorProps {
  index: number;
  question: QuestionCreateInput;
  onUpdate: (updates: Partial<QuestionCreateInput>) => void;
  onRemove: () => void;
}

const questionTypes = [
  { value: 'text', label: 'Texte', icon: FileText },
  { value: 'true_false', label: 'Vrai/Faux', icon: HelpCircle },
  { value: 'image', label: 'Image', icon: Image },
  { value: 'gif', label: 'GIF', icon: Image },
  { value: 'audio', label: 'Audio', icon: Headphones },
  { value: 'character_guess', label: 'Devine', icon: Users },
  { value: 'impostor', label: 'Intrus', icon: Search },
];

export function QuestionEditor({ index, question, onUpdate, onRemove }: QuestionEditorProps) {
  const updateAnswer = (answerIndex: number, updates: { answer_text?: string; is_correct?: boolean }) => {
    const newAnswers = question.answers.map((a, i) =>
      i === answerIndex ? { ...a, ...updates } : a
    );
    onUpdate({ answers: newAnswers });
  };

  const addAnswer = () => {
    if (question.answers.length >= 6) return;
    onUpdate({ answers: [...question.answers, { answer_text: '', is_correct: false }] });
  };

  const removeAnswer = (answerIndex: number) => {
    if (question.answers.length <= 2) return;
    onUpdate({ answers: question.answers.filter((_, i) => i !== answerIndex) });
  };

  const isTrueFalse = question.question_type === 'true_false';

  return (
    <Card className="border-dark-border bg-dark-card">
      <CardContent className="p-4 space-y-4">
        <div className="flex items-center gap-2">
          <GripVertical className="h-5 w-5 text-muted-foreground cursor-grab" />
          <span className="font-display text-lg">Question {index + 1}</span>
          <div className="flex-1" />
          <Button variant="ghost" size="icon" onClick={onRemove} className="text-red-400 hover:text-red-300">
            <Trash2 className="h-4 w-4" />
          </Button>
        </div>

        {/* Type de question */}
        <div className="flex flex-wrap gap-2">
          {questionTypes.map((type) => (
            <button
              key={type.value}
              onClick={() => onUpdate({ question_type: type.value as any })}
              className={cn(
                'flex items-center gap-1 px-3 py-1.5 rounded-md text-sm transition-colors',
                question.question_type === type.value
                  ? 'bg-brand text-white'
                  : 'bg-dark-surface text-muted-foreground hover:text-white'
              )}
            >
              <type.icon className="h-3.5 w-3.5" />
              {type.label}
            </button>
          ))}
        </div>

        {/* Texte de la question */}
        <textarea
          value={question.question_text}
          onChange={(e) => onUpdate({ question_text: e.target.value })}
          placeholder="Ta question..."
          className="w-full p-3 rounded-lg bg-dark-surface border border-dark-border text-white placeholder:text-muted-foreground focus:border-brand focus:outline-none resize-none"
          rows={2}
        />

        {/* Média (si applicable) */}
        {['image', 'image_shadow', 'gif', 'audio'].includes(question.question_type) && (
          <div className="flex gap-2">
            <input
              type="url"
              placeholder="URL du média..."
              className="flex-1 p-2 rounded-lg bg-dark-surface border border-dark-border text-sm text-white placeholder:text-muted-foreground focus:border-brand focus:outline-none"
              onChange={(e) => onUpdate({ media_url: e.target.value })}
            />
          </div>
        )}

        {/* Éditeur Character Guess - 4 indices pour 1 seul personnage */}
        {question.question_type === 'character_guess' && (
          <div className="space-y-4">
            {/* Sélecteur de mode Image/Texte */}
            <div className="flex items-center gap-2">
              <span className="text-sm text-muted-foreground">Mode :</span>
              <div className="flex gap-2">
                <button
                  onClick={() => onUpdate({ character_guess_mode: 'image' })}
                  className={cn(
                    'flex items-center gap-1 px-3 py-1.5 rounded-md text-sm transition-colors',
                    (question.character_guess_mode || 'image') === 'image'
                      ? 'bg-brand text-white'
                      : 'bg-dark-surface text-muted-foreground hover:text-white'
                  )}
                >
                  <Image className="h-3.5 w-3.5" />
                  Image
                </button>
                <button
                  onClick={() => onUpdate({ character_guess_mode: 'text' })}
                  className={cn(
                    'flex items-center gap-1 px-3 py-1.5 rounded-md text-sm transition-colors',
                    question.character_guess_mode === 'text'
                      ? 'bg-brand text-white'
                      : 'bg-dark-surface text-muted-foreground hover:text-white'
                  )}
                >
                  <FileText className="h-3.5 w-3.5" />
                  Texte
                </button>
              </div>
            </div>
            <p className="text-sm text-muted-foreground">
              Ajoute 4 indices du même personnage/anime. Les joueurs devront deviner le nom.
            </p>

            {/* 4 indices pour le même personnage */}
            <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
              {[0, 1, 2, 3].map((i) => {
                const characters = question.character_guess_data?.characters || [];
                const char = characters[i] || { image_url: '', clue: '', answer: '' };
                const updateClue = (field: 'image_url' | 'clue', value: string) => {
                  const newCharacters = [...characters];
                  newCharacters[i] = { ...newCharacters[i], [field]: value };
                  onUpdate({
                    character_guess_data: {
                      characters: newCharacters as CharacterGuessItem[],
                      mode: question.character_guess_mode || 'image',
                    },
                  });
                };

                return (
                  <div key={i} className="p-3 rounded-lg bg-dark-surface border border-dark-border space-y-2">
                    <span className="text-xs font-medium text-brand">Indice {i + 1}</span>
                    {(question.character_guess_mode || 'image') === 'image' ? (
                      <input
                        type="url"
                        value={char.image_url || ''}
                        onChange={(e) => updateClue('image_url', e.target.value)}
                        placeholder="URL de l'image..."
                        className="w-full p-2 rounded-lg bg-dark-card border border-dark-border text-sm text-white placeholder:text-muted-foreground focus:border-brand focus:outline-none"
                      />
                    ) : (
                      <textarea
                        value={char.clue || ''}
                        onChange={(e) => updateClue('clue', e.target.value)}
                        placeholder="Indice textuel..."
                        className="w-full p-2 rounded-lg bg-dark-card border border-dark-border text-sm text-white placeholder:text-muted-foreground focus:border-brand focus:outline-none resize-none"
                        rows={2}
                      />
                    )}
                  </div>
                );
              })}
            </div>

            {/* Un seul champ pour le nom du personnage */}
            <div className="p-4 rounded-lg bg-dark-surface border-2 border-brand/30 space-y-2">
              <span className="text-sm font-medium text-brand">🎯 Réponse : Nom du personnage/anime</span>
              <input
                type="text"
                value={question.character_guess_data?.characters?.[0]?.answer || ''}
                onChange={(e) => {
                  const val = e.target.value.toUpperCase();
                  const characters = question.character_guess_data?.characters || [];
                  const newCharacters = characters.map((c, idx) => ({
                    ...c,
                    answer: idx === 0 ? val : '',
                  }));
                  if (newCharacters.length === 0) {
                    newCharacters.push({ image_url: '', clue: '', answer: val });
                  }
                  onUpdate({
                    character_guess_data: {
                      characters: newCharacters as CharacterGuessItem[],
                      mode: question.character_guess_mode || 'image',
                    },
                  });
                }}
                placeholder="Ex: NARUTO UZUMAKI"
                className="w-full p-3 rounded-lg bg-dark-card border border-dark-border text-lg text-white placeholder:text-muted-foreground focus:border-brand focus:outline-none uppercase"
              />
              {question.character_guess_data?.characters?.[0]?.answer && (
                <div className="flex gap-1 justify-center pt-2">
                  {question.character_guess_data.characters[0].answer.split('').map((_, li) => (
                    <span key={li} className="w-8 h-10 flex items-center justify-center bg-dark-card border-2 border-brand/50 text-sm font-bold font-mono text-brand">
                      {'_'}
                    </span>
                  ))}
                </div>
              )}
            </div>
          </div>
        )}

        {/* Éditeur Impostor - Trouve l'intrus */}
        {question.question_type === 'impostor' && (
          <div className="space-y-4">
            <p className="text-sm text-muted-foreground">
              Ajoute 4 éléments (3 similaires + 1 intrus). Le joueur doit cliquer sur l'intrus.
            </p>
            <div className="flex items-center gap-2">
              <span className="text-sm text-muted-foreground">Mode :</span>
              <div className="flex gap-2">
                <button
                  onClick={() => onUpdate({ character_guess_mode: 'image' })}
                  className={cn(
                    'flex items-center gap-1 px-3 py-1.5 rounded-md text-sm transition-colors',
                    (question.character_guess_mode || 'image') === 'image'
                      ? 'bg-brand text-white'
                      : 'bg-dark-surface text-muted-foreground hover:text-white'
                  )}
                >
                  <Image className="h-3.5 w-3.5" />
                  Images
                </button>
                <button
                  onClick={() => onUpdate({ character_guess_mode: 'text' })}
                  className={cn(
                    'flex items-center gap-1 px-3 py-1.5 rounded-md text-sm transition-colors',
                    question.character_guess_mode === 'text'
                      ? 'bg-brand text-white'
                      : 'bg-dark-surface text-muted-foreground hover:text-white'
                  )}
                >
                  <FileText className="h-3.5 w-3.5" />
                  Texte
                </button>
              </div>
            </div>
            <div className="grid grid-cols-2 gap-3">
              {[0, 1, 2, 3].map((i) => {
                const items = question.find_odd_data?.items || [];
                const item = items[i] || { type: 'text', content: '' };
                const isOdd = question.find_odd_data?.odd_index === i;
                const updateItem = (field: 'content', value: string) => {
                  const newItems = [...items];
                  newItems[i] = { ...newItems[i], [field]: value } as FindOddItem;
                  onUpdate({
                    find_odd_data: {
                      items: newItems as FindOddItem[],
                      odd_index: question.find_odd_data?.odd_index ?? 0,
                    },
                  });
                };

                return (
                  <div key={i} className={cn(
                    'p-3 rounded-lg border space-y-2',
                    isOdd ? 'bg-red-500/10 border-red-500/50' : 'bg-dark-surface border-dark-border'
                  )}>
                    <div className="flex items-center justify-between">
                      <span className={cn(
                        'text-xs font-medium',
                        isOdd ? 'text-red-400' : 'text-brand'
                      )}>
                        {isOdd ? '🔴 INTRUS' : `Élément ${i + 1}`}
                      </span>
                      <button
                        onClick={() => onUpdate({
                          find_odd_data: {
                            items: items as FindOddItem[],
                            odd_index: i,
                          },
                        })}
                        className={cn(
                          'px-2 py-0.5 rounded text-xs transition-colors',
                          isOdd
                            ? 'bg-red-500 text-white'
                            : 'bg-dark-card text-muted-foreground hover:text-white'
                        )}
                      >
                        {isOdd ? 'Clique ici !' : 'Marquer intrus'}
                      </button>
                    </div>
                    {(question.character_guess_mode || 'image') === 'image' ? (
                      <input
                        type="url"
                        value={item.content}
                        onChange={(e) => updateItem('content', e.target.value)}
                        placeholder="URL de l'image..."
                        className="w-full p-2 rounded-lg bg-dark-card border border-dark-border text-sm text-white placeholder:text-muted-foreground focus:border-brand focus:outline-none"
                      />
                    ) : (
                      <input
                        type="text"
                        value={item.content}
                        onChange={(e) => updateItem('content', e.target.value)}
                        placeholder="Ex: Chapeau de paille..."
                        className="w-full p-2 rounded-lg bg-dark-card border border-dark-border text-sm text-white placeholder:text-muted-foreground focus:border-brand focus:outline-none"
                      />
                    )}
                  </div>
                );
              })}
            </div>
          </div>
        )}

        {/* Timer (caché pour character_guess et impostor) */}
        {!['character_guess', 'impostor'].includes(question.question_type) && (
          <div className="flex items-center gap-3">
            <span className="text-sm text-muted-foreground">Temps:</span>
            <input
              type="range"
              min={3}
              max={60}
              value={question.time_limit_seconds}
              onChange={(e) => onUpdate({ time_limit_seconds: parseInt(e.target.value) })}
              className="flex-1"
            />
            <span className="text-sm font-medium w-12">{question.time_limit_seconds}s</span>
          </div>
        )}

        {/* Réponses (caché pour character_guess et impostor) */}
        {!['character_guess', 'impostor'].includes(question.question_type) && (
        <div className="space-y-2">
          <span className="text-sm font-medium">Réponses</span>
          {isTrueFalse ? (
            <div className="grid grid-cols-2 gap-3">
              {['Vrai', 'Faux'].map((label, i) => (
                <button
                  key={i}
                  onClick={() => onUpdate({ answers: [
                    { answer_text: 'Vrai', is_correct: i === 0 },
                    { answer_text: 'Faux', is_correct: i === 1 },
                  ]})}
                  className={cn(
                    'p-3 rounded-lg border-2 transition-colors',
                    question.answers[i]?.is_correct
                      ? 'border-green-500 bg-green-500/10 text-green-400'
                      : 'border-dark-border bg-dark-surface text-muted-foreground'
                  )}
                >
                  {label}
                </button>
              ))}
            </div>
          ) : (
            <>
              {question.answers.map((answer, i) => (
                <div key={i} className="flex items-center gap-2">
                  <button
                    onClick={() => {
                      // Radio: une seule bonne réponse
                      const newAnswers = question.answers.map((a, idx) => ({
                        ...a,
                        is_correct: idx === i,
                      }));
                      onUpdate({ answers: newAnswers });
                    }}
                    className={cn(
                      'h-5 w-5 rounded-full border-2 shrink-0 transition-colors',
                      answer.is_correct
                        ? 'border-green-500 bg-green-500'
                        : 'border-dark-border'
                    )}
                  />
                  <input
                    type="text"
                    value={answer.answer_text}
                    onChange={(e) => updateAnswer(i, { answer_text: e.target.value })}
                    placeholder={`Réponse ${i + 1}`}
                    className="flex-1 p-2 rounded-lg bg-dark-surface border border-dark-border text-sm text-white placeholder:text-muted-foreground focus:border-brand focus:outline-none"
                  />
                  {question.answers.length > 2 && (
                    <Button variant="ghost" size="icon" onClick={() => removeAnswer(i)} className="h-8 w-8 text-red-400">
                      <Trash2 className="h-3 w-3" />
                    </Button>
                  )}
                </div>
              ))}
              {question.answers.length < 6 && (
                <Button variant="ghost" onClick={addAnswer} className="text-sm text-muted-foreground">
                  + Ajouter une réponse
                </Button>
              )}
            </>
          )}
        </div>
        )}
      </CardContent>
    </Card>
  );
}

export type { QuestionEditorProps };

