export type Screen = "home" | "searching" | "room" | "join-room" | "game" | "opponent-left";

export type Dish = {
  id: string;
  name: string;
  aliases: string[];
};

export type Player = {
  name: string;
  avatar: ChefAvatar;
  score: number;
  guess?: string;
  isYou?: boolean;
};

export type ChefAvatar = {
  color: "paprika" | "herb" | "mustard" | "aubergine" | "blue" | "rose";
  style: number;
};

export type Catalog = {
  dishes: Dish[];
};

export type MatchPlayer = {
  name: string;
  avatar: ChefAvatar;
  score: number;
  latestGuess?: string;
  isBot?: boolean;
};

export type MatchSnapshot = {
  id: string;
  token?: string;
  roomCode?: string;
  phase: "waiting" | "playing" | "result" | "finished";
  round: number;
  totalClueCount: number;
  clues: string[];
  /** Progressive assistance. The server withholds these until the player has
   *  earned them: family at half the clues, cuisine at three-quarters. */
  family?: string;
  cuisine?: string;
  player: MatchPlayer;
  opponent: MatchPlayer;
  nextRevealAt?: string;
  deadlineAt?: string;
  lockUntil?: string;
  answer?: Dish;
  roundWinner?: "player" | "opponent" | "none";
  matchWinner?: "player" | "opponent" | "none";
};
