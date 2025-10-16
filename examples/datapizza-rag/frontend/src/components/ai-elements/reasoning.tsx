"use client";

import { useControllableState } from "@radix-ui/react-use-controllable-state";
import { Collapsible } from "@/components/ui/collapsible";
import { cn } from "@/lib/utils";
import React, { createContext, memo, useEffect, useState } from "react";

type ReasoningContextValue = {
  isStreaming: boolean;
  isOpen: boolean;
  setIsOpen: (open: boolean) => void;
  duration: number;
};

const ReasoningContext = createContext<ReasoningContextValue | null>(null);


export type ReasoningProps = {
  className?: string;
  children?: React.ReactNode;
  isStreaming?: boolean;
  open?: boolean;
  defaultOpen?: boolean;
  onOpenChange?: (open: boolean) => void;
  duration?: number;
};

const AUTO_CLOSE_DELAY = 1000;
const MS_IN_S = 1000;

export const Reasoning = memo(
  ({
    className,
    isStreaming = false,
    open,
    defaultOpen = true,
    onOpenChange,
    duration: durationProp,
    children,
    ...props
  }: ReasoningProps) => {
    const [isOpen, setIsOpen] = useControllableState({
      prop: open,
      defaultProp: defaultOpen,
      onChange: onOpenChange,
    });
    const [duration, setDuration] = useControllableState({
      prop: durationProp,
      defaultProp: 0,
    });

    const [hasAutoClosed, setHasAutoClosed] = useState(false);
    const [startTime, setStartTime] = useState<number | null>(null);

    // Track duration when streaming starts and ends
    useEffect(() => {
      if (isStreaming) {
        if (startTime === null) {
          setStartTime(Date.now());
        }
      } else if (startTime !== null) {
        setDuration(Math.ceil((Date.now() - startTime) / MS_IN_S));
        setStartTime(null);
      }
    }, [isStreaming, startTime, setDuration]);

    // Auto-open when streaming starts, auto-close when streaming ends (once only)
    useEffect(() => {
      if (defaultOpen && !isStreaming && isOpen && !hasAutoClosed) {
        // Add a small delay before closing to allow user to see the content
        const timer = setTimeout(() => {
          setIsOpen(false);
          setHasAutoClosed(true);
        }, AUTO_CLOSE_DELAY);

        return () => clearTimeout(timer);
      }
    }, [isStreaming, isOpen, defaultOpen, setIsOpen, hasAutoClosed]);

    const handleOpenChange = (newOpen: boolean) => {
      setIsOpen(newOpen);
    };

    return (
      <ReasoningContext.Provider
        value={{ isStreaming, isOpen, setIsOpen, duration }}
      >
        <div className={cn("not-prose mb-4", className)}>
          <Collapsible
            onOpenChange={handleOpenChange}
            open={isOpen}
            {...props}
          >
            {children}
          </Collapsible>
        </div>
      </ReasoningContext.Provider>
    );
  }
);


Reasoning.displayName = "Reasoning";
