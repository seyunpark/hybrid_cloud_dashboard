interface LoadingSpinnerProps {
  size?: 'sm' | 'md' | 'lg';
  message?: string;
}

const sizeMap = {
  sm: 'h-4 w-4',
  md: 'h-8 w-8',
  lg: 'h-12 w-12',
};

export function LoadingSpinner({ size = 'md', message }: LoadingSpinnerProps) {
  return (
    <div className="flex flex-col items-center justify-center gap-2 p-4">
      <div
        className={`${sizeMap[size]} animate-spin rounded-full border-2 border-gray-300 border-t-blue-600`}
      />
      {message && <p className="text-sm text-gray-500">{message}</p>}
    </div>
  );
}
