'use client';

// ============================================================
// FORMULAIRE CRÉATION DE QUIZ
// ============================================================

import { useState } from 'react';
import { useRouter } from '../../../node_modules/next/navigation';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Progress } from '@/components/ui/progress';
import { QuestionEditor } from './QuestionEditor';
import { AIQuizImporter } from './AIQuizImporter';
import { createQuiz, updateQuiz } from '@/lib/actions/quiz';

import { QuizCreateInput, QuestionCreateInput, FindOddItem } from '@/types';
import { CATEGORY_LIST, SUBCATEGORY_LIST, QUIZ_MIN_QUESTIONS } from '@/lib/constants';
import { toast } from '@/lib/hooks/useToast';

import { Plus, Save, Send, ChevronLeft, ChevronRight, Image, Sparkles, Users, Trash2 } from 'lucide-react';
import { cn } from '@/lib/utils';
import { deleteQuestion } from '@/lib/actions/quiz';

const steps = ['Informations', 'Questions', 'Révision'];

interface QuizCreatorFormProps {
  quizId?: string;
  initialData?: {
    title: string;
    description: string | null;
    category: string;
    subcategory: string;
    series: string;
    thumbnail_url: string | null;
    duration_seconds?: number | null;
    duration_mode?: string | null;
    questions?: any[];
  };
}

export function QuizCreatorForm({ quizId, initialData }: QuizCreatorFormProps) {
  const router = useRouter();
  const [step, setStep] = useState(0);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const isEditing = !!quizId;

  // Étape 1: Informations
  const [title, setTitle] = useState(initialData?.title || '');
  const [description, setDescription] = useState(initialData?.description || '');
  const [category, setCategory] = useState(initialData?.category || '');
  const [subcategory, setSubcategory] = useState(initialData?.subcategory || '');
  const [series, setSeries] = useState(initialData?.series || '');
  const [thumbnailUrl, setThumbnailUrl] = useState(initialData?.thumbnail_url || '');

  // Étape 2: Questions
  const [questions, setQuestions] = useState<QuestionCreateInput[]>(() => {
    if (!initialData?.questions || initialData.questions.length === 0) return [];
    
    return initialData.questions.map((q: any) => ({
      question_text: q.question_text || '',
      question_type: q.question_type || 'text',
      time_limit_seconds: q.time_limit_seconds || 30,
      media_url: q.media_url || undefined,
      media_public_id: q.media_public_id || undefined,
      answers: (q.answers || []).map((a: any) => ({
        answer_text: a.answer_text || '',
        is_correct: a.is_correct || false,
      })),
    }));
  });

  console.log('Questions chargées:', questions.length, 'initialData:', initialData?.questions?.length);

  // Stocker les IDs des questions existantes (pour suppression en BDD)
  const [existingQuestionIds, setExistingQuestionIds] = useState<string[]>(() => {
    if (!initialData?.questions) return [];
    return initialData.questions.map((q: any) => q.id).filter(Boolean);
  });
  
  const [globalTimeLimit, setGlobalTimeLimit] = useState(initialData?.duration_seconds || 30);

  const addQuestion = (type: string = 'text') => {
    const newQuestion: QuestionCreateInput = {
      question_text: '',
      question_type: type as any,
      time_limit_seconds: globalTimeLimit,
      answers: type === 'character_guess' 
        ? [] 
        : [
            { answer_text: '', is_correct: false },
            { answer_text: '', is_correct: false },
          ],
    };

    if (type === 'character_guess') {
      newQuestion.guess_data = {
        clues: [
          { type: 'text', content: '' },
          { type: 'text', content: '' },
          { type: 'text', content: '' },
          { type: 'text', content: '' },
        ],
        answer: '',
      };
    }

    setQuestions((prev) => [...prev, newQuestion]);
  };

  const updateGlobalTimeLimit = (newTime: number) => {
    setGlobalTimeLimit(newTime);
    setQuestions((prev) => prev.map((q) => ({ ...q, time_limit_seconds: newTime })));
  };

  const updateQuestion = (index: number, updates: Partial<QuestionCreateInput>) => {
    setQuestions((prev) =>
      prev.map((q, i) => (i === index ? { ...q, ...updates } : q))
    );
  };

  const removeQuestion = async (index: number) => {
    const questionId = existingQuestionIds[index];
    
    if (isEditing && questionId) {
      try {
        await deleteQuestion(questionId);
        toast({ title: 'Question supprimée' });
      } catch (error: any) {
        toast({ title: 'Erreur', description: error.message, variant: 'destructive' });
        return;
      }
    }
    
    setQuestions((prev) => prev.filter((_, i) => i !== index));
    setExistingQuestionIds((prev) => prev.filter((_, i) => i !== index));
  };

  const handleAIImport = (importedQuestions: QuestionCreateInput[]) => {
    setQuestions((prev) => [...prev, ...importedQuestions]);
  };

  const handleSubmit = async (publish: boolean = true) => {
    if (questions.length < QUIZ_MIN_QUESTIONS) {
      toast({ title: 'Erreur', description: 'Minimum ' + QUIZ_MIN_QUESTIONS + ' questions requises', variant: 'destructive' });
      return;
    }

    setIsSubmitting(true);

    const quizData: QuizCreateInput = {
      title,
      description: description || undefined,
      category,
      subcategory,
      series,
      thumbnail_url: thumbnailUrl || undefined,
      duration_seconds: globalTimeLimit,
      duration_mode: 'per_question',
      questions: questions.map(q => ({
        ...q,
        time_limit_seconds: globalTimeLimit,
      })),
    };

    let result;

    if (isEditing && quizId) {
      result = await updateQuiz(quizId, quizData);
      if (result.success) {
        toast({ title: 'Succès', description: 'Quiz modifié avec succès !', variant: 'default' });
        router.push('/quiz/' + quizId);
      } else {
        toast({ title: 'Erreur', description: result.error || 'Erreur inconnue', variant: 'destructive' });
      }
    } else {
      result = await createQuiz(quizData);
      if (result.success && result.data) {
        toast({ title: 'Succès', description: 'Quiz créé avec succès !', variant: 'default' });
        router.push('/quiz/' + result.data.id);
      } else {
        toast({ title: 'Erreur', description: result.error || 'Erreur inconnue', variant: 'destructive' });
      }
    }

    setIsSubmitting(false);
  };

  const canProceed = () => {
  if (step === 0) {
    return title.length >= 5 && category && subcategory && series;
  }
  if (step === 1) {
    return questions.length >= QUIZ_MIN_QUESTIONS && questions.every((q) => {
      if (q.question_type === 'character_guess') {
        const chars = q.character_guess_data?.characters || [];
        const hasAtLeastOneClue = chars.some((c: any) =>
          (c.image_url && c.image_url.trim() !== '') ||
          (c.clue && c.clue.trim() !== '')
        );
        const hasAnswer = chars[0]?.answer && chars[0].answer.trim() !== '';
        return hasAtLeastOneClue && hasAnswer;
      }
      const hasCorrect = q.answers.some((a: any) => a.is_correct);
      const allFilled = q.question_text && q.answers.every((a: any) => a.answer_text);
      return hasCorrect && allFilled;
    });
  }
  return true;
};

  return (
    <div className='space-y-6'>
      {/* Stepper */}
      <div className='flex items-center gap-4'>
        {steps.map((s, i) => (
          <div key={s} className='flex items-center gap-2'>
            <div className={cn(
              'h-8 w-8 rounded-full flex items-center justify-center text-sm font-bold',
              i === step ? 'bg-brand text-white' :
              i < step ? 'bg-green-500 text-white' :
              'bg-dark-surface text-muted-foreground'
            )}>
              {i < step ? '✓' : i + 1}
            </div>
            <span className={cn(
              'text-sm',
              i === step ? 'text-white' : 'text-muted-foreground'
            )}>
              {s}
            </span>
            {i < steps.length - 1 && <div className='w-8 h-px bg-dark-border' />}
          </div>
        ))}
      </div>

      <Progress value={((step + 1) / steps.length) * 100} className='h-2' />

      {/* Étape 1: Informations */}
      {step === 0 && (
        <Card className='border-dark-border bg-dark-card'>
          <CardContent className='p-6 space-y-4'>
            <div>
              <label className='text-sm font-medium mb-1 block'>Titre *</label>
              <input
                type='text'
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                placeholder='Ex: Quiz Naruto - Les techniques'
                className='w-full p-3 rounded-lg bg-dark-surface border border-dark-border text-white placeholder:text-muted-foreground focus:border-brand focus:outline-none'
                maxLength={100}
              />
              <div className='text-xs text-muted-foreground mt-1'>{title.length}/100</div>
            </div>

            <div>
              <label className='text-sm font-medium mb-1 block'>Description</label>
              <textarea
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder='Décris ton quiz...'
                className='w-full p-3 rounded-lg bg-dark-surface border border-dark-border text-white placeholder:text-muted-foreground focus:border-brand focus:outline-none resize-none'
                rows={3}
                maxLength={500}
              />
              <div className='text-xs text-muted-foreground mt-1'>{description.length}/500</div>
            </div>

            <div className='grid grid-cols-2 gap-4'>
              <div>
                <label className='text-sm font-medium mb-1 block'>Catégorie *</label>
                <select
                  value={category}
                  onChange={(e) => { setCategory(e.target.value); setSubcategory(''); }}
                  className='w-full p-3 rounded-lg bg-dark-surface border border-dark-border text-white focus:border-brand focus:outline-none'
                >
                  <option value=''>Choisir...</option>
                  {CATEGORY_LIST.map((c) => (
                    <option key={c} value={c}>{c}</option>
                  ))}
                </select>
              </div>
              <div>
                <label className='text-sm font-medium mb-1 block'>Sous-catégorie *</label>
                <select
                  value={subcategory}
                  onChange={(e) => setSubcategory(e.target.value)}
                  disabled={!category}
                  className='w-full p-3 rounded-lg bg-dark-surface border border-dark-border text-white focus:border-brand focus:outline-none disabled:opacity-50'
                >
                  <option value=''>Choisir...</option>
                  {category && SUBCATEGORY_LIST[category]?.map((s) => (
                    <option key={s} value={s}>{s}</option>
                  ))}
                </select>
              </div>
            </div>

            <div>
              <label className='text-sm font-medium mb-1 block'>Série *</label>
              <input
                type='text'
                value={series}
                onChange={(e) => setSeries(e.target.value)}
                placeholder='Ex: Naruto, One Piece, Attack on Titan...'
                className='w-full p-3 rounded-lg bg-dark-surface border border-dark-border text-white placeholder:text-muted-foreground focus:border-brand focus:outline-none'
              />
            </div>

            <div>
              <label className='text-sm font-medium mb-1 block'>Thumbnail URL</label>
              <div className='flex gap-2'>
                <input
                  type='url'
                  value={thumbnailUrl}
                  onChange={(e) => setThumbnailUrl(e.target.value)}
                  placeholder='URL de l-image...'
                  className='flex-1 p-3 rounded-lg bg-dark-surface border border-dark-border text-white placeholder:text-muted-foreground focus:border-brand focus:outline-none'
                />
                <Button variant='outline' size='icon'>
                  <Image className='h-4 w-4' />
                </Button>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Étape 2: Questions */}
      {step === 1 && (
        <div className='space-y-4'>
          {/* Timer global */}
          <Card className='border-dark-border bg-dark-card'>
            <CardContent className='p-4'>
              <div className='flex items-center gap-4'>
                <span className='text-sm font-medium'>Temps par question (global):</span>
                <input
                  type='range'
                  min={3}
                  max={60}
                  value={globalTimeLimit}
                  onChange={(e) => updateGlobalTimeLimit(parseInt(e.target.value))}
                  className='flex-1'
                />
                <span className='text-sm font-medium w-12'>{globalTimeLimit}s</span>
              </div>
            </CardContent>
          </Card>

          <div className='flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3'>
            <span className='text-sm text-muted-foreground'>
              {questions.length} question{questions.length > 1 ? 's' : ''} (min {QUIZ_MIN_QUESTIONS})
            </span>
            <div className='flex flex-wrap items-center gap-2'>
              <AIQuizImporter onImport={handleAIImport} globalTimeLimit={globalTimeLimit} />
              <Button onClick={() => addQuestion('text')} className='gap-2'>
                <Plus className='h-4 w-4' /> Ajouter
              </Button>
            </div>
          </div>

          {questions.map((q, i) => (
            <QuestionEditor
              key={i}
              index={i}
              question={q}
              onUpdate={(updates) => updateQuestion(i, updates)}
              onRemove={() => removeQuestion(i)}
            />
          ))}

          {questions.length === 0 && (
            <Card className='border-dashed border-dark-border bg-dark-card/50'>
              <CardContent className='p-8 text-center'>
                <Plus className='h-8 w-8 mx-auto mb-2 text-muted-foreground' />
                <p className='text-muted-foreground'>Ajoute ta première question ou importe depuis une IA</p>
              </CardContent>
            </Card>
          )}
        </div>
      )}

      {/* Étape 3: Révision */}
      {step === 2 && (
        <Card className='border-dark-border bg-dark-card'>
          <CardContent className='p-6 space-y-4'>
            <h3 className='font-display text-xl tracking-wider'>RÉCAPITULATIF</h3>
            <div className='space-y-2 text-sm'>
              <p><span className='text-muted-foreground'>Titre:</span> {title}</p>
              <p><span className='text-muted-foreground'>Catégorie:</span> {category} / {subcategory}</p>
              <p><span className='text-muted-foreground'>Série:</span> {series}</p>
              <p><span className='text-muted-foreground'>Questions:</span> {questions.length}</p>
            </div>
            <div className='flex gap-4 pt-4'>
              <Button onClick={() => handleSubmit(true)} disabled={isSubmitting} className='gap-2'>
                <Send className='h-4 w-4' />
                {isSubmitting ? 'Publication...' : 'Publier maintenant'}
              </Button>
              <Button variant='outline' onClick={() => handleSubmit(false)} disabled={isSubmitting} className='gap-2'>
                <Save className='h-4 w-4' /> Sauvegarder brouillon
              </Button>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Navigation */}
      <div className='flex items-center justify-between pt-4'>
        <Button
          variant='ghost'
          onClick={() => setStep((s) => Math.max(0, s - 1))}
          disabled={step === 0}
          className='gap-2'
        >
          <ChevronLeft className='h-4 w-4' /> Précédent
        </Button>
        <Button
          onClick={() => setStep((s) => Math.min(steps.length - 1, s + 1))}
          disabled={step === steps.length - 1 || !canProceed()}
          className='gap-2'
        >
          Suivant <ChevronRight className='h-4 w-4' />
        </Button>
      </div>
    </div>
  );
}
