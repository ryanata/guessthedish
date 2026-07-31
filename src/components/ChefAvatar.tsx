import type { ChefAvatar as ChefAvatarType } from "../types";

const colors: Record<ChefAvatarType["color"], string> = {
  paprika: "#d84a32",
  herb: "#47725a",
  mustard: "#c3922e",
  aubergine: "#725068",
  blue: "#397387",
  rose: "#b85f68",
};

export function ChefAvatar({ avatar }: { avatar: ChefAvatarType }) {
  const accent = colors[avatar.color];
  const style = avatar.style % 4;

  return (
    <svg className="chef-avatar" viewBox="0 0 56 56" role="img" aria-label={`${avatar.color} chef portrait, style ${style + 1}`}>
      <rect width="56" height="56" fill="#efe5d4" />
      <path d="M10 56V45c0-10 7-16 18-16s18 6 18 16v11Z" fill={accent} stroke="#25221d" strokeWidth="2" />
      <path d="M20 27v7l8 5 8-5v-7" fill="#fff8ea" stroke="#25221d" strokeWidth="2" />
      <path d="M17 15c0-8 5-13 11-13s11 5 11 13v9c0 8-5 13-11 13S17 32 17 24Z" fill="#fff8ea" stroke="#25221d" strokeWidth="2" />
      <path d="M14 14V9c0-3 2-5 5-5 1-3 4-4 7-3 2-2 6-1 7 2 5-1 8 2 8 6v5Z" fill="#fff8ea" stroke="#25221d" strokeWidth="2" strokeLinejoin="round" />
      <path d="M15 13c8-2 18-2 26 0v5c-8-2-18-2-26 0Z" fill="#fff8ea" stroke="#25221d" strokeWidth="2" />
      <path d="M22 22h1m10 0h1" stroke="#25221d" strokeWidth="2" strokeLinecap="round" />
      {style === 0 && <path d="M23 29c3 2 7 2 10 0" fill="none" stroke="#25221d" strokeWidth="2" strokeLinecap="round" />}
      {style === 1 && <path d="M21 29c2-4 5-4 7-1 2-3 5-3 7 1-3 0-5 1-7 3-2-2-4-3-7-3Z" fill="#25221d" />}
      {style === 2 && <><circle cx="22.5" cy="22" r="4" fill="none" stroke="#25221d" strokeWidth="1.5" /><circle cx="33.5" cy="22" r="4" fill="none" stroke="#25221d" strokeWidth="1.5" /><path d="M27 22h2" stroke="#25221d" /></>}
      {style === 3 && <><path d="M19 19l6-1m6 0 6 1" stroke="#25221d" strokeWidth="2" strokeLinecap="round" /><path d="M24 29h8" stroke="#25221d" strokeWidth="2" strokeLinecap="round" /></>}
    </svg>
  );
}
