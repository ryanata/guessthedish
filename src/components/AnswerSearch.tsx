import { useEffect, useRef, useState } from "react";
import type { Dish } from "../types";

type AnswerSearchProps = {
  disabled: boolean;
  locked: boolean;
  dishes: Dish[];
  onSubmit: (dish: Dish) => void;
};

export function AnswerSearch({ disabled, locked, dishes, onSubmit }: AnswerSearchProps) {
  const inputRef = useRef<HTMLInputElement>(null);
  const [query, setQuery] = useState("");
  const [selectedIndex, setSelectedIndex] = useState(0);
  const [tooQuick, setTooQuick] = useState(false);
  const normalizedQuery = query.trim().toLowerCase();
  const results = normalizedQuery.length < 2
    ? []
    : dishes.filter((dish) =>
        [dish.name, ...dish.aliases].some((name) =>
          name.toLowerCase().includes(normalizedQuery),
        ),
      ).slice(0, 5);

  useEffect(() => {
    if (!disabled) inputRef.current?.focus();
  }, [disabled]);

  function submit(dish: Dish, fromKeyboard = false) {
    if (disabled || locked) {
      if (fromKeyboard) {
        setTooQuick(true);
        window.setTimeout(() => setTooQuick(false), 700);
      }
      return;
    }
    onSubmit(dish);
    setQuery("");
    setSelectedIndex(0);
  }

  return (
    <div className={`answer-search ${locked ? "answer-search-locked" : ""}`}>
      <label htmlFor="answer">Your answer</label>
      <input
        ref={inputRef}
        id="answer"
        value={query}
        disabled={disabled}
        autoComplete="off"
        placeholder="Start typing a dish..."
        onChange={(event) => {
          setQuery(event.target.value);
          setSelectedIndex(0);
        }}
        onKeyDown={(event) => {
          if (event.key === "ArrowDown") {
            event.preventDefault();
            setSelectedIndex((index) => Math.min(index + 1, results.length - 1));
          }
          if (event.key === "ArrowUp") {
            event.preventDefault();
            setSelectedIndex((index) => Math.max(index - 1, 0));
          }
          if (event.key === "Enter" && results[selectedIndex]) {
            event.preventDefault();
            submit(results[selectedIndex], true);
          }
        }}
      />
      {results.length > 0 && !disabled && (
        <div className="answer-results" role="listbox" aria-label="Dish suggestions">
          {results.map((dish, index) => (
            <button
              className={index === selectedIndex ? "answer-result-selected" : ""}
              key={dish.id}
              onClick={() => submit(dish)}
              role="option"
              aria-selected={index === selectedIndex}
            >
              {dish.name}
            </button>
          ))}
        </div>
      )}
      {tooQuick && <span className="too-quick" role="status">Too quick - hold on</span>}
    </div>
  );
}
