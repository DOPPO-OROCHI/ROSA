import { useLayoutEffect, useRef, useState } from "react";

type Props = {
  text: string;
  className?: string;
  maxFontSize: number;
  minFontSize: number;
};

export function AutoFitText({ text, className, maxFontSize, minFontSize }: Props) {
  const containerRef = useRef<HTMLSpanElement | null>(null);
  const textRef = useRef<HTMLSpanElement | null>(null);
  const [fontSize, setFontSize] = useState(maxFontSize);

  useLayoutEffect(() => {
    const container = containerRef.current;
    const textNode = textRef.current;

    if (!container || !textNode) {
      return;
    }

    let nextFontSize = maxFontSize;
    textNode.style.fontSize = `${nextFontSize}px`;

    while (nextFontSize > minFontSize && textNode.scrollWidth > container.clientWidth) {
      nextFontSize -= 0.5;
      textNode.style.fontSize = `${nextFontSize}px`;
    }

    setFontSize(nextFontSize);
  }, [maxFontSize, minFontSize, text]);

  return (
    <span ref={containerRef} className={className}>
      <span ref={textRef} className="auto-fit-text__content" style={{ fontSize: `${fontSize}px` }}>
        {text}
      </span>
    </span>
  );
}
