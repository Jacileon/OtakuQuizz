'use client';

// ============================================================
// QUIZ ENGINE - Composant client (CRITIQUE)
// RÈGLES: Pas de is_correct exposé, pas de feedback pendant quiz
// ============================================================

import { useState, useCallback, useEffect, useRef } from 'react';
import { useRouter } from 'next/navigation';
import { cn } from '@/lib/utils';

import { QuestionClient, PlayerAnswerDraft, FindOddItem } from '@/types';
import { Button } from '@/components/ui/button';
import { Progress } from '@/components/ui/progress';
import { Card } from '@/components/ui/card';
import { LoadingSpinner } from '@/components/ui/LoadingSpinner';
import { QuestionDisplay } from './QuestionDisplay';
import { AnswerButton } from './AnswerButton';
import { Timer } from './Timer';
import { toast } from '@/lib/hooks/useToast';
import { ChevronRight, Send, RotateCcw, Edit2, ArrowLeft, Play } from 'lucide-react';

interface QuizEngineProps {
  quizId: string;
  sessionId: string;
  questions: QuestionClient[];
  isOfficial: boolean;
  quizTitle: string;
}

type QuizStatus = 'countdown' | 'playing' | 'submitting' | 'completed';

export function QuizEngine({ quizId, sessionId, questions, isOfficial, quizTitle }: QuizEngineProps) {
  const router = useRouter();
  const [countdown, setCountdown] = useState(3);
  const [currentIndex, setCurrentIndex] = useState(0);
  const [selectedAnswerId, setSelectedAnswerId] = useState<string | null>(null);
  const [answers, setAnswers] = useState<PlayerAnswerDraft[]>([]);
  const [questionStartTime, setQuestionStartTime] = useState(Date.now());
  const [quizStatus, setQuizStatus] = useState<QuizStatus>('countdown');
  const [quizResults, setQuizResults] = useState<any>(null);
  const [textInput, setTextInput] = useState('');
  const [selectedItemIndex, setSelectedItemIndex] = useState<number | null>(null);

  // Use ref to track answers without causing re-renders
  const answersRef = useRef<PlayerAnswerDraft[]>([]);
  
  const currentQuestion = questions[currentIndex];
  const totalQuestions = questions.length;
  const progress = totalQuestions > 0 ? ((currentIndex + 1) / totalQuestions) * 100 : 0;
  const isCharacterGuess = currentQuestion?.question_type === 'character_guess';
  const isImpostor = currentQuestion?.question_type === 'impostor';

  // Sync ref with state
  useEffect(() => {
    answersRef.current = answers;
  }, [answers]);

  // Reset text input on question change
  useEffect(() => {
    setTextInput('');
  }, [currentIndex]);

  // Gestion du countdown
  useEffect(() => {
    if (quizStatus !== 'countdown') return;
    
    if (countdown > 0) {
      const timer = setTimeout(() => {
        setCountdown(countdown - 1);
      }, 1000);
      return () => clearTimeout(timer);
    } else {
      // Countdown terminé, lancer le quiz
      setQuizStatus('playing');
      setQuestionStartTime(Date.now());
    }
  }, [countdown, quizStatus]);

  const handleAnswerSelect = (answerId: string) => {
    if (quizStatus !== 'playing') return;
    setSelectedAnswerId(answerId);
  };

  const handleItemSelect = (index: number) => {
    if (quizStatus !== 'playing') return;
    setSelectedItemIndex(index);
  };

  // Pour character_guess, trouver l'answerId correspondant au texte tapé
  const findMatchingAnswerId = (text: string): string | null => {
    if (!currentQuestion) return null;
    const normalizedInput = text.trim().toUpperCase();
    const correctAnswer = currentQuestion.character_guess_data?.characters?.[0]?.answer?.toUpperCase();
    if (correctAnswer && normalizedInput === correctAnswer) {
      // Trouver la première réponse correspondante dans answers
      const matchingAnswer = currentQuestion.answers.find(
        a => a.answer_text.toUpperCase() === correctAnswer
      );
      return matchingAnswer?.id || null;
    }
    return null;
  };

  const handleNext = useCallback(() => {
    if (!currentQuestion) return;
    // Vérifications selon le type
    if (isCharacterGuess) {
      if (!textInput.trim()) return;
    } else if (isImpostor) {
      if (selectedItemIndex === null) return;
    } else {
      if (!selectedAnswerId || quizStatus !== 'playing') return;
    }

    const timeTaken = Date.now() - questionStartTime;
    
    let answerId = selectedAnswerId;
    if (isCharacterGuess) {
      answerId = findMatchingAnswerId(textInput);
    } else if (isImpostor) {
      // Trouver l'answerId correspondant à l'index sélectionné (l'intrus)
      if (selectedItemIndex === currentQuestion.find_odd_data?.odd_index) {
        const oddItem = currentQuestion.find_odd_data?.items[selectedItemIndex];
        if (oddItem) {
          const matchingAnswer = currentQuestion.answers.find(
            a => a.answer_text.toUpperCase() === oddItem.content.toUpperCase()
          );
          answerId = matchingAnswer?.id || null;
        } else {
          answerId = null;
        }
      } else {
        answerId = null;
      }
    }

    const newAnswer: PlayerAnswerDraft = {
      questionId: currentQuestion.id,
      answerId: answerId,
      timeMs: Math.min(timeTaken, currentQuestion.time_limit_seconds * 1000),
    };

    const updatedAnswers = [...answersRef.current, newAnswer];
    setAnswers(updatedAnswers);
    answersRef.current = updatedAnswers;
    setSelectedAnswerId(null);
    setSelectedItemIndex(null);
    setTextInput('');

    if (currentIndex < totalQuestions - 1) {
      setCurrentIndex((prev) => prev + 1);
      setQuestionStartTime(Date.now());
    } else {
      handleSubmit(updatedAnswers);
    }
  }, [selectedAnswerId, selectedItemIndex, textInput, isCharacterGuess, isImpostor, currentQuestion, currentIndex, totalQuestions, questionStartTime, quizStatus]);

  const handleSubmit = async (finalAnswers: PlayerAnswerDraft[]) => {
    setQuizStatus('submitting');

    try {
      const response = await fetch('/api/quiz/submit', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          sessionId,
          answers: finalAnswers,
        }),
      });

      if (!response.ok) {
        throw new Error('Erreur lors de la soumission');
      }

      const results = await response.json();
      console.log('Quiz results:', results);
      
      setQuizResults(results);
      setQuizStatus('completed');
    } catch (error) {
      console.error('Submit error:', error);
      toast({
        title: 'Erreur',
        description: 'Impossible de soumettre les réponses. Réessaie.',
        variant: 'destructive',
      });
      setQuizStatus('playing');
    }
  };

  const handleTimeUp = useCallback(() => {
    if (quizStatus !== 'playing' || !currentQuestion) return;
    
    // Si temps écoulé, soumettre sans réponse (null)
    const timeTaken = currentQuestion.time_limit_seconds * 1000;
    const newAnswer: PlayerAnswerDraft = {
      questionId: currentQuestion.id,
      answerId: null,
      timeMs: timeTaken,
    };

    const updatedAnswers = [...answersRef.current, newAnswer];
    setAnswers(updatedAnswers);
    answersRef.current = updatedAnswers;
    setSelectedAnswerId(null);
    setSelectedItemIndex(null);

    if (currentIndex < totalQuestions - 1) {
      setCurrentIndex((prev) => prev + 1);
      setQuestionStartTime(Date.now());
    } else {
      handleSubmit(updatedAnswers);
    }
  }, [currentQuestion, currentIndex, totalQuestions, quizStatus]);

  const handleRestart = () => {
    // Réinitialiser tous les états pour recommencer
    setCurrentIndex(0);
    setSelectedAnswerId(null);
    setSelectedItemIndex(null);
    setAnswers([]);
    answersRef.current = [];
    setQuestionStartTime(Date.now());
    setQuizStatus('countdown');
    setCountdown(3);
    setQuizResults(null);
  };

  const handleEditQuiz = () => {
    router.push(`/quiz/${quizId}/edit`);
  };

  const handleBackToQuizList = () => {
    router.push('/profil');
  };

  // État de countdown
  if (quizStatus === 'countdown') {
    return (
      <div className="min-h-screen flex flex-col items-center justify-center bg-dark">
        <div className="text-center space-y-6">
          <div className="relative">
            <div className={cn(
              "text-9xl font-bold text-brand transition-all duration-300",
              countdown === 0 ? "scale-150 opacity-100" : "scale-100 opacity-100"
            )}>
              {countdown === 0 ? 'GO!' : countdown}
            </div>
            {countdown > 0 && (
              <div className="absolute inset-0 flex items-center justify-center">
                <div className="w-32 h-32 rounded-full border-4 border-brand/20 animate-ping" />
              </div>
            )}
          </div>
          <p className="text-xl text-muted-foreground">
            {countdown === 0 ? "C'est parti!" : "Prépare-toi..."}
          </p>
          <div className="flex items-center gap-2 text-sm text-muted-foreground">
            <Play className="h-4 w-4" />
            <span>{quizTitle}</span>
          </div>
        </div>
      </div>
    );
  }

  // État de soumission
  if (quizStatus === 'submitting') {
    return (
      <div className="min-h-screen flex flex-col items-center justify-center bg-dark">
        <LoadingSpinner size="lg" />
        <p className="mt-4 text-lg text-muted-foreground animate-pulse">
          Calcul du score...
        </p>
      </div>
    );
  }

  // État de complétion - Afficher les résultats
  if (quizStatus === 'completed' && quizResults) {
    return (
      <div className="min-h-screen flex flex-col items-center justify-center bg-dark p-6">
        <div className="max-w-md w-full space-y-6">
          <div className="text-center">
            <h2 className="text-2xl font-bold mb-2">Quiz Terminé ! 🎉</h2>
            <p className="text-muted-foreground">
              Tu as répondu à toutes les questions.
            </p>
          </div>
          
          <Card className="border-dark-border bg-dark-card">
            <div className="p-6">
              <div className="text-center space-y-4">
                <div>
                  <p className="text-5xl font-bold text-brand mb-2">
                    {quizResults.correctCount ?? 0}
                  </p>
                  <p className="text-lg text-muted-foreground">
                    sur {totalQuestions} questions
                  </p>
                </div>
                <div className="pt-4 border-t border-dark-border">
                  <p className="text-2xl font-semibold">
                    {quizResults.accuracyRate ?? 0}%
                  </p>
                  <p className="text-sm text-muted-foreground">de bonnes réponses</p>
                </div>
                {quizResults.details && (
                  <div className="pt-4 text-sm text-muted-foreground">
                    <p>Temps moyen: {Math.round(quizResults.avgTimeMs / 1000)}s par question</p>
                  </div>
                )}
              </div>
            </div>
          </Card>

          <div className="grid gap-3">
            <Button 
              onClick={handleRestart}
              className="w-full gap-2"
              variant="default"
            >
              <RotateCcw className="h-4 w-4" />
              Recommencer le quiz
            </Button>
            
            <Button 
              onClick={handleEditQuiz}
              className="w-full gap-2"
              variant="secondary"
            >
              <Edit2 className="h-4 w-4" />
              Modifier ce quiz
            </Button>
            
            <Button 
              onClick={handleBackToQuizList}
              className="w-full gap-2"
              variant="ghost"
            >
              <ArrowLeft className="h-4 w-4" />
              Retour à la liste des quiz
            </Button>
          </div>
        </div>
      </div>
    );
  }

  // État de jeu - Afficher la question
  return (
    <div className="min-h-screen bg-dark p-4 lg:p-8">
      <div className="max-w-3xl mx-auto space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="font-display text-xl tracking-wider">{quizTitle}</h1>
            <p className="text-sm text-muted-foreground">
              Question {currentIndex + 1} / {totalQuestions}
            </p>
          </div>
          {isOfficial && (
            <div className="flex items-center gap-2 px-3 py-1 rounded-full bg-brand/10 border border-brand/20 text-brand text-sm">
              <div className="h-2 w-2 rounded-full bg-brand animate-pulse" />
              OFFICIEL
            </div>
          )}
        </div>

        {/* Progress */}
        <Progress value={progress} className="h-2" />

        {/* Timer */}
        {currentQuestion && (
        <Timer
          key={`timer-${currentIndex}`}
          duration={currentQuestion.time_limit_seconds || 30}
          onTimeUp={handleTimeUp}
          isActive={quizStatus === 'playing'}
        />
      )}

        {/* Question */}
        {currentQuestion && (
        <Card className="border-dark-border bg-dark-card">
          <div className="p-6">
            <QuestionDisplay question={currentQuestion} />
          </div>
        </Card>
      )}

        {/* Affichage des éléments pour impostor */}
        {isImpostor && currentQuestion && (
          <div className="grid grid-cols-2 gap-3">
            {(currentQuestion.find_odd_data?.items || []).map((item, i) => (
              <button
                key={i}
                onClick={() => handleItemSelect(i)}
                disabled={quizStatus !== 'playing'}
                className={cn(
                  'p-4 rounded-lg border-2 transition-all duration-200',
                  selectedItemIndex === i
                    ? 'border-brand bg-brand/10 shadow-lg shadow-brand/10'
                    : 'border-dark-border bg-dark-card hover:border-brand/50 hover:bg-dark-surface'
                )}
              >
                {item.type === 'image' && item.content ? (
                  <img
                    src={item.content}
                    alt={`Élément ${i + 1}`}
                    className="w-full h-32 object-cover rounded-lg"
                  />
                ) : (
                  <p className="text-lg font-medium text-center">{item.content || '?'}</p>
                )}
              </button>
            ))}
          </div>
        )}

        {/* Answers ou Input pour character_guess (pas pour impostor) */}
        {!isImpostor && (isCharacterGuess ? (
          <div className="space-y-3">
            <p className="text-sm text-muted-foreground text-center">
              Tape le nom du personnage/anime en majuscules
            </p>
            <input
              type="text"
              value={textInput}
              onChange={(e) => setTextInput(e.target.value.toUpperCase())}
              placeholder="Tape ta réponse ici..."
              className="w-full p-4 rounded-lg bg-dark-card border-2 border-dark-border text-center text-xl font-bold text-white placeholder:text-muted-foreground focus:border-brand focus:outline-none uppercase"
              disabled={quizStatus !== 'playing'}
              autoFocus
              onKeyDown={(e) => {
                if (e.key === 'Enter' && textInput.trim()) {
                  handleNext();
                }
              }}
            />
            {textInput && (
              <div className="flex gap-1 justify-center">
                {textInput.split('').map((char, i) => (
                  <span
                    key={i}
                    className="w-10 h-12 flex items-center justify-center bg-dark-surface border-2 border-brand/50 text-lg font-bold font-mono text-brand"
                  >
                    {char}
                  </span>
                ))}
              </div>
            )}
          </div>
        ) : (
          <div className="grid gap-3">
            {currentQuestion?.answers.map((answer) => (
              <AnswerButton
                key={answer.id}
                answer={answer}
                isSelected={selectedAnswerId === answer.id}
                onSelect={() => handleAnswerSelect(answer.id)}
                disabled={quizStatus !== 'playing'}
              />
            ))}
          </div>
        ))}

        {/* Navigation */}
        <div className="flex items-center justify-end pt-4">
          <Button
            onClick={handleNext}
            disabled={
              isCharacterGuess
                ? !textInput.trim()
                : isImpostor
                ? selectedItemIndex === null
                : !selectedAnswerId
            }
            className="gap-2"
          >
            {currentIndex === totalQuestions - 1 ? (
              <>
                <Send className="h-4 w-4" /> Valider
              </>
            ) : (
              <>
                Suivant <ChevronRight className="h-4 w-4" />
              </>
            )}
          </Button>
        </div>
      </div>
    </div>
  );
}
