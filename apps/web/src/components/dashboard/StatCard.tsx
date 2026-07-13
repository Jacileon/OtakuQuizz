import { cn } from '@/lib/utils';
import { Card, CardContent } from '@/components/ui/card';

interface StatCardProps {
  icon: React.ReactNode;
  label: string;
  value: string;
  color: 'brand' | 'accent' | 'green' | 'purple';
}

const colorMap = {
  brand: 'text-brand bg-brand/10',
  accent: 'text-accent bg-accent/10',
  green: 'text-green-400 bg-green-400/10',
  purple: 'text-purple-400 bg-purple-400/10',
};

export function StatCard({ icon, label, value, color }: StatCardProps) {
  return (
    <Card className="border-dark-border bg-dark-card/50">
      <CardContent className="p-4 flex items-center gap-4">
        <div className={cn('h-10 w-10 rounded-lg flex items-center justify-center', colorMap[color])}>
          {icon}
        </div>
        <div>
          <div className="font-display text-2xl">{value}</div>
          <div className="text-xs text-muted-foreground">{label}</div>
        </div>
      </CardContent>
    </Card>
  );
}

