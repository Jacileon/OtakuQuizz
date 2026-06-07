'use client';

import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { QuestionCreateInput, CharacterGuessItem, FindOddItem } from '@/types';
import { toast } from '@/lib/hooks/useToast';
import { Sparkles, Check, AlertCircle, Loader2 } from 'lucide-react';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
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
  character_guess_data?: { characters: CharacterGuessItem[] };
  find_odd_data?: { items: FindOddItem[]; odd_index: number };
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

    // ============ PARSER DEVINE ============
    const devineBlocks = cleanedText.split(/DEVINE\s*\n/i).filter(b => b.trim());
    for (const block of devineBlocks) {
      const lines = block.trim().split('\n').filter(l => l.trim());
      const clues: string[] = [];
      let answer = '';
      for (const line of lines) {
        const clueMatch = line.match(/Indice\s+\d+\s*:\s*(.+)/i);
        const answerMatch = line.match(/Réponse\s*:\s*(.+)/i);
        if (clueMatch) clues.push(clueMatch[1].trim());
        else if (answerMatch) answer = answerMatch[1].trim().toUpperCase();
      }
      if (clues.length >= 2 && answer) {
        while (clues.length < 4) clues.push('?');
        questions.push({
          question_text: 'Devine ce personnage/anime !',
          question_type: 'character_guess',
          answers: [{ answer_text: answer, is_correct: true }],
          time_limit_seconds: globalTimeLimit,
          character_guess_data: { characters: clues.map(clue => ({ image_url: undefined, answer, clue })) },
        });
      }
    }

    // ============ PARSER IMPOSTEUR ============
    const imposteurBlocks = cleanedText.split(/IMPOSTEUR\s*-\s*/i).filter(b => b.trim());
    for (const block of imposteurBlocks) {
      const lines = block.trim().split('\n').filter(l => l.trim());
      if (lines.length < 5) continue;
      const items: string[] = [];
      let oddIndex = 0;
      for (let i = 0; i < 4 && i + 1 < lines.length; i++) {
        const m = lines[i + 1]?.match(/^\d+\.\s*(.+)/);
        if (m) {
          const isOdd = m[1].includes('(INTRUS)');
          items.push(m[1].replace(/\(INTRUS\)/gi, '').trim());
          if (isOdd) oddIndex = i;
        }
      }
      if (items.length === 4) {
        questions.push({
          question_text: "Trouve l'intrus !",
          question_type: 'impostor',
          answers: [{ answer_text: items[oddIndex], is_correct: true }],
          time_limit_seconds: globalTimeLimit,
          find_odd_data: { items: items.map(item => ({ type: 'text' as const, content: item })), odd_index: oddIndex },
        });
      }
    }

    // ============ PARSER QCM & VRAI/FAUX ============
    const questionBlocks = cleanedText.split(/\n(?=\d+[).]\s*)/).filter(b => b.trim());
    for (const block of questionBlocks) {
      const trimmedBlock = block.trim();
      const tfMatch = trimmedBlock.match(/(?:\d+[).]\s*)?(.+?)\.?\s*(Vrai|Faux|True|False)\s*$/i);
      if (tfMatch) {
        questions.push({
          question_text: tfMatch[1].replace(/^\d+[).]\s*/, '').trim(),
          question_type: 'true_false',
          answers: [
            { answer_text: 'Vrai', is_correct: tfMatch[2].toLowerCase() === 'vrai' || tfMatch[2].toLowerCase() === 'true' },
            { answer_text: 'Faux', is_correct: tfMatch[2].toLowerCase() === 'faux' || tfMatch[2].toLowerCase() === 'false' },
          ],
          time_limit_seconds: globalTimeLimit,
        });
        continue;
      }
      const lines = trimmedBlock.split('\n').filter(l => l.trim());
      if (lines.length < 2) continue;
      const questionText = lines[0].replace(/^\d+[).]\s*/, '').trim();
      const answers: { answer_text: string; is_correct: boolean }[] = [];
      for (let i = 1; i < lines.length; i++) {
        const m = lines[i].trim().match(/^[A-Za-z][).]\s*(.+)/);
        if (m) {
          let t = m[1].trim();
          const ok = t.includes('✓') || t.includes('✔') || t.toLowerCase().includes('(correct)') || t.toLowerCase().includes('[correct]');
          t = t.replace(/[✓✔]/g, '').replace(/\(correct\)/gi, '').replace(/\[correct\]/gi, '').trim();
          if (t) answers.push({ answer_text: t, is_correct: ok });
        }
      }
      if (answers.length >= 2 && answers.some(a => a.is_correct)) {
        questions.push({ question_text: questionText, question_type: 'text', answers, time_limit_seconds: globalTimeLimit });
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
        setParseError('Aucune question détectée. Vérifie le format.');
      } else {
        setPreviewQuestions(parsed);
        toast({ title: 'Parsing réussi', description: `${parsed.length} question(s) détectée(s)`, variant: 'default' });
      }
    } catch {
      setParseError('Erreur lors du parsing.');
    } finally {
      setIsParsing(false);
    }
  };

  const handleImport = () => {
    if (!previewQuestions.length) return;
    onImport(previewQuestions.map(q => {
      const base: any = {
        question_text: q.question_text,
        question_type: q.question_type,
        time_limit_seconds: q.time_limit_seconds,
        answers: q.answers,
        character_guess_data: q.character_guess_data,
        find_odd_data: q.find_odd_data,
      };
      if (q.question_type === 'character_guess') {
        base.character_guess_mode = 'text';
      }
      return base;
    }));
    toast({ title: 'Importé !', description: `${previewQuestions.length} question(s) ajoutée(s)`, variant: 'default' });
    setIsOpen(false);
    setInputText('');
    setPreviewQuestions([]);
    setParseError(null);
  };

  const handleClose = () => { setIsOpen(false); setInputText(''); setPreviewQuestions([]); setParseError(null); };

  const typeLabel = (t: string) => t === 'true_false' ? 'V/F' : t === 'character_guess' ? 'Devine' : t === 'impostor' ? 'Imposteur' : 'QCM';

  return (
    <Dialog open={isOpen} onOpenChange={o => { if (!o) handleClose(); else setIsOpen(true); }}>
      <DialogTrigger asChild>
        <Button variant='outline' className='gap-2'><Sparkles className='h-4 w-4' />Importer depuis IA</Button>
      </DialogTrigger>
      <DialogContent className='max-w-2xl max-h-[80vh] overflow-y-auto bg-dark-card border-dark-border'>
        <DialogHeader>
          <DialogTitle>Importer depuis une IA</DialogTitle>
          <DialogDescription>Colle le texte généré par ChatGPT, Claude ou une autre IA. Les questions seront parsées automatiquement.</DialogDescription>
        </DialogHeader>
        <div className='space-y-4 mt-4'>
          <textarea
            value={inputText}
            onChange={e => setInputText(e.target.value)}
            placeholder={'Formats acceptés :\n\nQCM :\n1. Question ?\nA) Reponse ✓\nB) Autre\n\nVrai/Faux :\n2. Naruto est Hokage. Vrai\n\nDEVINE :\nDEVINE\nIndice 1: Cheveux orange\nIndice 2: Rasengan\nIndice 3: Fils du 4e Hokage\nRéponse: NARUTO\n\nIMPOSTEUR :\nIMPOSTEUR - Naruto\n1. Sasuke\n2. Sakura\n3. Kakashi\n4. Luffy (INTRUS)'}
            className='w-full h-48 p-3 rounded-lg bg-dark-surface border border-dark-border text-white placeholder:text-muted-foreground focus:border-brand focus:outline-none resize-none text-sm'
          />
          <Button onClick={handleParse} disabled={!inputText.trim() || isParsing} className='w-full gap-2'>
            {isParsing ? <Loader2 className='h-4 w-4 animate-spin' /> : <Sparkles className='h-4 w-4' />}
            {isParsing ? 'Analyse...' : 'Analyser le texte'}
          </Button>
          {parseError && (
            <div className='flex items-start gap-2 p-3 rounded-lg bg-red-500/10 border border-red-500/30 text-red-400 text-sm'>
              <AlertCircle className='h-4 w-4 mt-0.5 shrink-0' /><p>{parseError}</p>
            </div>
          )}
          {previewQuestions.length > 0 && (
            <div className='space-y-3'>
              <p className='text-sm font-medium text-green-400 flex items-center gap-2'><Check className='h-4 w-4' />{previewQuestions.length} question(s) détectée(s)</p>
              <div className='space-y-2 max-h-48 overflow-y-auto pr-1'>
                {previewQuestions.map((q, i) => (
                  <div key={i} className='p-3 rounded-lg bg-dark-surface border border-dark-border text-sm'>
                    <div className='flex items-center justify-between mb-1'>
                      <p className='font-medium text-white'>{i + 1}. {q.question_text}</p>
                      <span className={cn('text-xs px-2 py-0.5 rounded',
                        q.question_type === 'impostor' ? 'bg-red-500/20 text-red-400' :
                        q.question_type === 'character_guess' ? 'bg-purple-500/20 text-purple-400' :
                        q.question_type === 'true_false' ? 'bg-blue-500/20 text-blue-400' : 'bg-green-500/20 text-green-400'
                      )}>{typeLabel(q.question_type)}</span>
                    </div>
                    {q.question_type === 'character_guess' && q.character_guess_data && (
                      <div className='mt-1'>{q.character_guess_data.characters.map((c, j) => <p key={j} className='text-xs text-muted-foreground'>Indice {j + 1}: {c.clue}</p>)}<p className='text-xs text-purple-400 font-medium mt-1'>Réponse: {q.answers[0]?.answer_text}</p></div>
                    )}
                    {q.question_type === 'impostor' && q.find_odd_data && (
                      <div className='mt-1'>{q.find_odd_data.items.map((item, j) => <p key={j} className={cn('text-xs', j === q.find_odd_data!.odd_index ? 'text-red-400 font-medium' : 'text-muted-foreground')}>{j + 1}. {item.content}{j === q.find_odd_data!.odd_index && ' (INTRUS)'}</p>)}</div>
                    )}
                    {(q.question_type === 'text' || q.question_type === 'true_false') && (
                      <div className='mt-1'>{q.answers.map((a, j) => <p key={j} className={cn('text-xs', a.is_correct ? 'text-green-400 font-medium' : 'text-muted-foreground')}>{a.is_correct ? '✓' : '○'} {a.answer_text}</p>)}</div>
                    )}
                  </div>
                ))}
              </div>
              <Button onClick={handleImport} className='w-full gap-2'><Check className='h-4 w-4' />Importer {previewQuestions.length} question(s)</Button>
            </div>
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}