type ChefIllustrationProps = {
  color: "red" | "green";
  facing?: "left" | "right";
};

export function ChefIllustration({ color, facing = "right" }: ChefIllustrationProps) {
  const accent = color === "red" ? "#D84A32" : "#47725A";

  return (
    <svg
      className={`chef-svg ${facing === "left" ? "chef-svg-flipped" : ""}`}
      viewBox="0 0 320 390"
      role="img"
      aria-label={`${color === "red" ? "Paprika" : "Herb"} chef`}
    >
      <g stroke="#25221D" strokeWidth="6" strokeLinecap="round" strokeLinejoin="round">
        <path fill={accent} d="M73 387v-75c0-62 41-101 97-101s97 39 97 101v75Z" />
        <path fill="#FFF8EA" d="M126 205v40l44 27 43-28v-42" />
        <path fill="#FFF8EA" d="M105 123c0-43 27-72 65-72s65 29 65 72v51c0 44-29 75-65 75s-65-31-65-75Z" />
        <path d="M130 145c9-7 19-7 28-1m24 0c9-7 19-6 28 2" fill="none" />
        {color === "red" ? (
          <>
            <circle cx="143" cy="160" r="18" fill="none" />
            <circle cx="197" cy="160" r="18" fill="none" />
            <path d="M161 159h18" fill="none" />
          </>
        ) : null}
        <path d="M144 160h1m51 0h1m-27 2-5 23h11" fill="none" />
        {color === "red" ? (
          <path fill="#25221D" d="M138 207c8-17 23-21 32-8 10-13 25-9 32 8-9-3-17-2-24 5-5 5-11 5-16 0-7-7-15-8-24-5Z" />
        ) : (
          <path d="M145 208c15 10 34 10 49 0" fill="none" />
        )}
        <path d="M170 272v115m-44-142 44 27-27 36-38-48m108-16-43 28 26 36 38-48" fill="none" />
        <path fill="#FFF8EA" d="M94 108V72c0-20 14-34 34-34 7-19 23-29 41-29 19 0 35 11 42 30 20 0 35 15 35 35v34Z" />
        <path d="M119 91V55m26 38V36m27 57V29m27 65V47m25 50V66" fill="none" strokeWidth="2.5" />
        <path fill="#FFF8EA" d="M96 99c24-7 49-10 74-10s50 3 74 10v27c-24-6-49-9-74-9s-50 3-74 9Z" />
        <path d="M103 112c44-10 90-10 134 0" fill="none" strokeWidth="2.5" />
      </g>
    </svg>
  );
}
