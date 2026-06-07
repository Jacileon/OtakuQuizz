import { LoadingSpinner } from '@/components/ui/LoadingSpinner';

export default function LoadingPage() {
  return (
    <div className="min-h-screen bg-dark flex items-center justify-center">
      <LoadingSpinner size="lg" />
    </div>
  );
}

