import { useEffect, useState } from 'react';
import type { ImgHTMLAttributes, ReactNode } from 'react';
import { assetUrl } from '../api/client';

interface Props extends Omit<ImgHTMLAttributes<HTMLImageElement>, 'src'> {
  src?: string | null;
  fallback?: ReactNode;
}

// Renders uploaded/remote images with a stable fallback so a 404 never leaves
// the browser's broken-image alt text on the page.
export default function SafeImage({ src, alt, className, fallback, ...rest }: Props) {
  const resolved = assetUrl(src);
  const [failed, setFailed] = useState(false);

  useEffect(() => {
    setFailed(false);
  }, [resolved]);

  if (!resolved || failed) {
    return (
      <>
        {fallback ?? (
          <div className={`img-fallback ${className || ''}`} role="img" aria-label={alt}>
            <span>粤</span>
          </div>
        )}
      </>
    );
  }

  return (
    <img
      {...rest}
      className={className}
      src={resolved}
      alt={alt}
      onError={() => setFailed(true)}
    />
  );
}
