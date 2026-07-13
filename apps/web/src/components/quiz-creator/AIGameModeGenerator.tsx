'use client';

// ============================================================
// IMPORTATEUR DE QUIZ DEPUIS UNE IA
// Parse le texte copié depuis ChatGPT, Claude, etc.
// ============================================================

import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { QuestionCreateInput, CharacterGuessItem } from '@/types';
import { toast } from '@/lib/hooks/useToast';
import { Sparkles, Check, AlertCircle, Loader2 } from 'lucide-react';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog';
import { cn } from '@/lib/utils';

interface AIQuizImporterProps {
  onImport: (questions: QuestionCreateInput[]) => void;
  globalTimeLimit?: number;
}

interface ParsedQuestion {
  question_text: string;
  question_type: 'text' | 'true_false' | 'character_guess' | 'impostor';
  answers: { answer_text: string; is_correct: boolean }[];
  time_limit_seconds: number;
  character_guess_data?: {
    characters: CharacterGuessItem[];
  };
  find_odd_data?: {
    items: { type: 'text'; content: string }[];
    odd_index: number;
  };
}

export function AIQuizImporter({ onImport, globalTimeLimit = 30 }: AIQuizImporterProps) {
  const [isOpen, setIsOpen] = useState(false);
  const [inputText, setInputText] = useState('');
  const [previewQuestions, setPreviewQuestions] = useState<ParsedQuestion[]>([]);
  const [isParsing, setIsParsing] = useState(false);
  const [parseError, setParseError] = useState<string | null>(null);

  const parseQuizText = (text: string): ParsedQuestion[] => {
    const questions: ParsedQuestion[] = [];
    const cleanedText = text.replace(/\r\n/g, '\n').trim();

    // ── PARSER DEVINE (character_guess) ──────────────────────────
    // Format attendu :
    // DEVINE
    // Indice 1: Il est blond
    // Indice 2: Il porte un bandeau
    // Indice 3: Son rêve est de devenir Hokage
    // Indice 4: Il a un renard en lui
    // Réponse: Naruto
    const devineSegments = cleanedText.split(/(?=DEVINE\s*\n)/gi).filter((s) => s.trim());
    for (const segment of devineSegments) {
      if (!/^DEVINE\s*\n/i.test(segment.trim())) continue;
      const lines = segment.trim().split('\n').filter((l) => l.trim());
      const clueLines = lines.filter((l) => /^indice\s*\d+\s*:/i.test(l));
      const answerLine = lines.find((l) => /^réponse\s*:/i.test(l));
      if (!answerLine || clueLines.length === 0) continue;

      const answer = answerLine.replace(/^réponse\s*:\s*/i, '').trim();
      const clues = clueLines.map((l) => ({
        type: 'text' as const,
        content: l.replace(/^indice\s*\d+\s*:\s*/i, '').trim(),
      }));

      questions.push({
        question_text: 'Qui est-ce ?',
        question_type: 'character_guess',
        answers: [],
        time_limit_seconds: globalTimeLimit,
        character_guess_data: {
          characters: [
            {
              answer,
              clues,
            } as unknown as CharacterGuessItem,
          ],
        },
      });
    }

    // ── PARSER IMPOSTEUR (find_odd) ───────────────────────────────
    // Format attendu :
    // IMPOSTEUR - Quel est l'intrus ?
    // - Naruto
    // - Sasuke
    // - Sakura
    // - Goku ✓
    const imposteurSegments = cleanedText.split(/(?=IMPOSTEUR\s*-)/gi).filter((s) => s.trim());
    for (const segment of imposteurSegments) {
      if (!/^IMPOSTEUR\s*-/i.test(segment.trim())) continue;
      const lines = segment.trim().split('\n').filter((l) => l.trim());
      const questionText = lines[0].replace(/^IMPOSTEUR\s*-\s*/i, '').trim() || "Quel est l'intrus ?";
      const itemLines = lines.slice(1).filter((l) => /^-\s+/.test(l));
      if (itemLines.length < 3) continue;

      let oddIndex = -1;
      const items = itemLines.map((l, idx) => {
        const isOdd = l.includes('✓') || l.includes('✔') || l.toLowerCase().includes('(intrus)');
        if (isOdd) oddIndex = idx;
        return {
          type: 'text' as const,
          content: l
            .replace(/^-\s+/, '')
            .replace(/[✓✔]/g, '')
            .replace(/\(intrus\)/gi, '')
            .trim(),
        };
      });

      if (oddIndex === -1) continue;

      questions.push({
        question_text: questionText,
        question_type: 'impostor',
        answers: [],
        time_limit_seconds: globalTimeLimit,
        find_odd_data: { items, odd_index: oddIndex },
      });
    }

    // ── PARSER QCM + VRAI/FAUX ────────────────────────────────────
    // Split par les numéros de question
    const questionBlocks = cleanedText.split(/\n(?=\d+[).]\s*)/).filter((block) => block.trim());

    for (const block of questionBlocks) {
      const trimmedBlock = block.trim();

      // Vrai/Faux
      const tfMatch = trimmedBlock.match(/(?:\d+[).]\s*)?(.+?)\.?\s*(Vrai|Faux|True|False)\s*$/i);
      if (tfMatch) {
        const questionText = tfMatch[1].replace(/^\d+[).]\s*/, '').trim();
        const answer = tfMatch[2];

        questions.push({
          question_text: questionText,
          question_type: 'true_false',
          answers: [
            {
              answer_text: 'Vrai',
              is_correct: answer.toLowerCase() === 'vrai' || answer.toLowerCase() === 'true',
            },
            {
              answer_text: 'Faux',
              is_correct: answer.toLowerCase() === 'faux' || answer.toLowerCase() === 'false',
            },
          ],
          time_limit_seconds: globalTimeLimit,
        });
        continue;
      }

      // QCM
      const lines = trimmedBlock.split('\n').filter((line) => line.trim());
      if (lines.length < 2) continue;

      const questionText = lines[0].replace(/^\d+[).]\s*/, '').trim();
      const answers: { answer_text: string; is_correct: boolean }[] = [];

      for (let i = 1; i < lines.length; i++) {
        const line = lines[i].trim();
        const answerMatch = line.match(/^[A-Za-z][).]\s*(.+)/);

        if (answerMatch) {
          let answerText = answerMatch[1].trim();
          const isCorrect =
            answerText.includes('✓') ||
            answerText.includes('✔') ||
            answerText.toLowerCase().includes('(correct)') ||
            answerText.toLowerCase().includes('[correct]');

          answerText = answerText
            .replace(/[✓✔]/g, '')
            .replace(/\(correct\)/gi, '')
            .replace(/\[correct\]/gi, '')
            .trim();

          if (answerText) {
            answers.push({ answer_text: answerText, is_correct: isCorrect });
          }
        }
      }

      if (answers.length >= 2 && answers.some((a) => a.is_correct)) {
        questions.push({
          question_text: questionText,
          question_type: 'text',
          answers,
          time_limit_seconds: globalTimeLimit,
        });
      }
    }

    return questions;
  };

  const handleParse = () => {
    setIsParsing(true);
    setParseError(null);

    try {
      const parsed = parseQuizText(inputText);

      if (parsed.length === 0) {
        setParseError(
          'Aucune question détectée. Vérifie le format : numérotation (1. ou 1)), options A) B) C), et marque ✓ ou (correct) sur la bonne réponse.'
        );
      } else {
        setPreviewQuestions(parsed);
        toast({
          title: 'Parsing réussi',
          description: `${parsed.length} question(s) détectée(s)`,
          variant: 'default',
        });
      }
    } catch (err) {
      setParseError('Erreur lors du parsing. Vérifie le format du texte.');
    } finally {
      setIsParsing(false);
    }
  };

  const handleImport = () => {
    if (previewQuestions.length === 0) return;

    onImport(previewQuestions as QuestionCreateInput[]);
    toast({
      title: 'Importé !',
      description: `${previewQuestions.length} question(s) ajoutée(s) au quiz`,
      variant: 'default',
    });
    setIsOpen(false);
    setInputText('');
    setPreviewQuestions([]);
    setParseError(null);
  };

  const handleClose = () => {
    setIsOpen(false);
    setInputText('');
    setPreviewQuestions([]);
    setParseError(null);
  };

  return (
    <Dialog open={isOpen} onOpenChange={(open) => { if (!open) handleClose(); else setIsOpen(true); }}>
      <DialogTrigger asChild>
        <Button variant='outline' className='gap-2'>
          <Sparkles className='h-4 w-4' />
          Importer depuis IA
        </Button>
      </DialogTrigger>

      <DialogContent className='max-w-2xl max-h-[80vh] overflow-y-auto bg-dark-card border-dark-border'>
        <DialogHeader>
          <DialogTitle>Importer depuis une IA</DialogTitle>
          <DialogDescription>
            Colle le texte généré par ChatGPT, Claude ou une autre IA. Les questions seront parsées automatiquement.
          </DialogDescription>
        </DialogHeader>

        <div className='space-y-4 mt-4'>
          {/* Zone de texte */}
          <textarea
            value={inputText}
            onChange={(e) => setInputText(e.target.value)}
            placeholder={`Formats supportés :\n\n── QCM ──\n1. Quel est le vrai nom de Naruto ?\nA) Naruto Namikaze ✓\nB) Naruto Sarutobi\n\n── DEVINE (character_guess) ──\nDEVINE\nIndice 1: Il est blond\nIndice 2: Il porte un bandeau\nIndice 3: Son rêve est de devenir Hokage\nIndice 4: Il a un renard en lui\nRéponse: Naruto\n\n── IMPOSTEUR (find_odd) ──\nIMPOSTEUR - Quel est l'intrus ?\n- Naruto\n- Sasuke\n- Sakura\n- Goku ✓`}
            className='w-full h-48 p-3 rounded-lg bg-dark-surface border border-dark-border text-white placeholder:text-muted-foreground focus:border-brand focus:outline-none resize-none text-sm'
          />

          {/* Bouton parser */}
          <Button
            onClick={handleParse}
            disabled={!inputText.trim() || isParsing}
            className='w-full gap-2'
          >
            {isParsing ? (
              <Loader2 className='h-4 w-4 animate-spin' />
            ) : (
              <Sparkles className='h-4 w-4' />
            )}
            {isParsing ? 'Analyse en cours...' : 'Analyser le texte'}
          </Button>

          {/* Erreur */}
          {parseError && (
            <div className='flex items-start gap-2 p-3 rounded-lg bg-red-500/10 border border-red-500/30 text-red-400 text-sm'>
              <AlertCircle className='h-4 w-4 mt-0.5 shrink-0' />
              <p>{parseError}</p>
            </div>
          )}

          {/* Aperçu */}
          {previewQuestions.length > 0 && (
            <div className='space-y-3'>
              <p className='text-sm font-medium text-green-400 flex items-center gap-2'>
                <Check className='h-4 w-4' />
                {previewQuestions.length} question(s) détectée(s)
              </p>

              <div className='space-y-2 max-h-48 overflow-y-auto pr-1'>
                {previewQuestions.map((q, i) => (
                  <div
                    key={i}
                    className='p-3 rounded-lg bg-dark-surface border border-dark-border text-sm'
                  >
                    <p className='font-medium text-white mb-1'>
                      {i + 1}. {q.question_text}
                      <span className='ml-2 text-xs text-muted-foreground'>({q.question_type})</span>
                    </p>
                    <div className='space-y-0.5'>
                      {q.question_type === 'impostor' && q.find_odd_data ? (
                        q.find_odd_data.items.map((item, j) => (
                          <p
                            key={j}
                            className={cn(
                              'text-xs',
                              j === q.find_odd_data!.odd_index
                                ? 'text-red-400 font-medium'
                                : 'text-muted-foreground'
                            )}
                          >
                            {j === q.find_odd_data!.odd_index ? '✗' : '○'} {item.content}
                          </p>
                        ))
                      ) : q.question_type === 'character_guess' && q.character_guess_data ? (
                        <p className='text-xs text-green-400'>
                          Réponse : {(q.character_guess_data.characters[0] as any)?.answer}
                          {' — '}
                          {(q.character_guess_data.characters[0] as any)?.clues?.length ?? 0} indice(s)
                        </p>
                      ) : (
                        q.answers.map((a, j) => (
                          <p
                            key={j}
                            className={cn(
                              'text-xs',
                              a.is_correct ? 'text-green-400 font-medium' : 'text-muted-foreground'
                            )}
                          >
                            {a.is_correct ? '✓' : '○'} {a.answer_text}
                          </p>
                        ))
                      )}
                    </div>
                  </div>
                ))}
              </div>

              {/* Bouton importer */}
              <Button onClick={handleImport} className='w-full gap-2'>
                <Check className='h-4 w-4' />
                Importer {previewQuestions.length} question(s)
              </Button>
            </div>
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}