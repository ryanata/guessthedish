# Guess the Dish

## Overview

Guess the Dish is a fast, one-versus-one web game. A dish's ingredients are
revealed one at a time, and players race to identify the dish by selecting an
answer from a fixed culinary catalog.

The central decision is whether to guess early with incomplete information or
wait for a more distinctive ingredient and risk losing the race.

## Terminology

- **Dish:** A canonical answer in the answer catalog, such as `Spaghetti
  Bolognese`.
- **Alias:** An accepted searchable name that resolves to a canonical dish,
  such as `spag bol`.
- **Puzzle:** A dish plus its curated, ordered ingredient clues.
- **Round:** One puzzle played until a correct answer or timeout.
- **Match:** A sequence of rounds until one player wins three rounds.
- **Room:** A temporary match lobby reached through an invite URL.

## Core Rules

- Matches are one versus one.
- The first player to win three rounds wins the match.
- A correct answer wins the round and awards one point.
- An unsolved round awards no point; the match continues with a new puzzle.
- A puzzle must not repeat within a match.
- The server determines which correct answer arrived first and ends the round
  immediately.
- There is no grace window or tied round.
- A match result remains visible until the player chooses to play again or
  leave.

## Clues And Timing

Puzzles may contain different numbers of displayed ingredient clues.

- Minimum puzzle size: four ingredients.
- Maximum puzzle size: twelve ingredients.
- Each clue shows an ingredient label without its quantity. Preparation may be
  included when it changes the ingredient's identity, such as `roasted red
  pepper`.
- Both players receive the same reveal order and server-controlled schedule.
- The first ingredient appears at `GO`; revealed ingredients remain visible.

Let:

- `N` be the number of ingredients in the puzzle.
- `K = ceil(N * 0.25)` be the number of early clues.

- Wait two seconds after each of the first `K` reveals and three seconds after
  each remaining reveal.
- After the final ingredient, leave a final three-second answer window.
- If nobody answers correctly, reveal the dish and award no point.

## Answering

Players do not submit arbitrary prose. They search and select from a static
catalog of valid dishes.

### Search

- Search begins after two typed characters.
- At most five results appear.
- Search covers canonical dish names and aliases.
- Results always display the canonical dish name.
- Selecting a result submits it immediately.
- The first result is highlighted by default. Arrow keys change the selection,
  and `Enter` submits it.
- Freeform text that does not resolve to a catalog entry cannot be submitted.
- Exact aliases rank ahead of partial matches. Conservative typo matching must
  still resolve to a catalog entry.

Example:

```text
Query: bolognese

Spaghetti Bolognese
Ragù alla Bolognese
```

### Visible Guesses

- Every submitted catalog answer appears in a compact speech bubble above the
  submitting player's identity.
- Both players see bubbles for their own guesses and their opponent's guesses.
- The bubble displays the selected dish's canonical name, even when the player
  found it through an alias.
- A bubble remains visible for two seconds. A new guess replaces it and
  restarts the display time.
- An incorrect guess remains visible while its 750 ms input lock runs.
- A correct guess remains visible through the transition into the round result.
- Bots use the same behavior. Bubbles never contain freeform text and are not a
  chat system.
- Bubbles must not obscure ingredients, timers, or answer controls.

### Wrong Answers

- An incorrect submission locks that player's answer control for 750 ms.
- The lock does not pause ingredient reveals.
- Previously rejected dishes remain selectable and may be submitted again.

## Puzzle Content

Each puzzle is curated rather than generated directly from an arbitrary source recipe.

- Stable puzzle identifier.
- Canonical dish identifier and display name.
- Searchable aliases.
- Ordered ingredient clues.
- Internal difficulty rating.
- Content status such as `draft`, `reviewed`, or `retired`.
- Optional cuisine, region, dish photograph, and attribution.

### Content Guidelines

- A clue set should converge on a conventional named dish rather than a broad
  category.
- Order ingredients from less identifying to more identifying. A highly
  revealing ingredient should not appear first unless the puzzle is easy.
- Salt, water, neutral oil, and similar pantry defaults may be omitted when
  they add no useful signal.
- Treat variants as separate answers only when their clues distinguish them.
- Aliases are equivalent names, not descriptions: `Spag bol` may resolve to
  `Spaghetti Bolognese`; `pasta with meat sauce` may not.
- Revise or retire frequently disputed puzzles.

## Play Modes

Players use guest names; accounts, friend lists, and social graphs are not
required.

### Quick Play

- **Quick Play** searches for a waiting player and starts when both seats are
  ready.
- If no player is available after three seconds, a bot fills the opposing seat.
- Bots use revealed clues, plausible reaction times, and configurable skill
  profiles rather than knowing the answer and waiting arbitrarily.
- Bots receive generated player-like names and are not explicitly labeled.

### Invite Room

- **Create Room** generates a temporary room with an unguessable invite URL the
  host can copy.
- The first guest to join claims the second seat; both players being ready
  starts the match automatically.
- A third visitor sees **Room full**.
- Invite rooms do not use bot fill and support rematches in the same room.

### Reconnection

- A disconnected player's seat remains reserved for 10 seconds while the round
  timeline continues.
- A returning client catches up to the current state.
- After 10 seconds, the remaining player wins by forfeit.

## Screens

### Home

- Guess the Dish wordmark.
- One-sentence explanation.
- Editable guest display name.
- Primary **Quick Play** action.
- Secondary **Create Room** action.

### Quick Play Matchmaking

- Visible searching state.
- Elapsed search feedback or lightweight animation.
- Cancel action.
- Transition directly into the match countdown when an opponent is found.

### Invite Room Lobby

- Room identifier or concise waiting label.
- Host and guest player slots.
- Shareable URL and **Copy Invite Link** action with confirmation feedback.
- Leave action.
- Clear states for expired, invalid, and full rooms.

### Game

- Both player names and connection states.
- Three score markers for each player.
- Current round number.
- Ordered ingredient reveal area with room for the puzzle's variable clue
  count.
- Clear indication of time until the next reveal.
- Persistent answer search field and autocomplete results.
- Wrong-answer lock feedback.

### Round Result

- Canonical dish name and winning player, or **No one got it**.
- Ingredient number visible when the winning answer was accepted.
- Full ordered clue sequence.
- Updated match score.
- Dish image when available and properly licensed.
- Brief automatic transition to the next round.

### Match Result

- Winner and final score.
- Round-by-round summary.
- **Play Again** for Quick Play or rematch readiness for invite rooms.
- Leave action.

## Responsive And Accessible Interaction

- The active game uses a mobile-first layout, expands on desktop, and never
  requires scrolling.
- On mobile, autocomplete opens upward so results remain visible above the
  software keyboard.
- Nonessential decoration collapses while the keyboard is open.
- Touch targets are at least 44 by 44 CSS pixels. The game is keyboard operable
  with visible focus states.
- Status never relies on color alone, and text maintains readable contrast.
- Announce new clues, submitted guesses, lockouts, round results, and
  connection changes to assistive technology without repeatedly stealing
  focus.
- Respect `prefers-reduced-motion`.

## Visual System

The visual direction is **Restaurant Ticket**: fast, tactile, warm, and
rooted in the rhythm of a working kitchen. It should feel like ink, paper, and
the practical artifacts of restaurant service rather than a generic trivia
game.

### Palette

| Role | Name | Value | Intended use |
| --- | --- | --- | --- |
| Primary surface | Paper | `#FFF8EA` | Cards, tickets, fields, and principal readable surfaces |
| Canvas | Kraft | `#EFE5D4` | Page background and quieter secondary areas |
| Primary text | Ink | `#25221D` | Text, strong borders, and high-emphasis controls |
| Action | Paprika | `#D84A32` | Primary actions, active clues, urgency, and focus |
| Success | Herb | `#47725A` | Correct answers, connection health, and confirmation |
| Supporting line | Ticket line | `#B8AA96` | Dividers, inactive borders, and structural rules |
| Muted text | Faded ink | `#756D61` | Secondary labels and supporting information |

Paprika and Herb communicate different meanings and should not be used as
interchangeable decoration. States must also use text, iconography, shape, or
another non-color signal.

### Typography

- **Display and utility face:** IBM Plex Mono.
- **Body and interface face:** DM Sans.
- Use IBM Plex Mono for ticket-like headings, timers, round identifiers,
  ingredient numbers, codes, and compact operational labels.
- Use DM Sans for player names, explanations, answer results, buttons, and
  longer interface copy.
- Headings may use uppercase and modest tracking when they resemble printed
  service labels; body copy should retain normal casing for readability.
- Tabular numerals should be used for countdowns and changing numeric values.

The fonts may be self-hosted or loaded as web fonts. The
interface must retain sensible system fallbacks and avoid layout shifts that
affect gameplay.

### Shape And Material

- Predominantly square corners or a very small radius, rather than soft app
  cards and pills.
- Strong ink borders for active controls; lighter ticket-line borders for
  secondary structure.
- Dashed rules may suggest perforation or separation where semantically
  useful.
- Subtle paper grain is permitted on large surfaces but must never reduce text
  clarity.
- Small rotations or imperfect registration may be used sparingly for static
  ticket details, never for inputs, timers, or dense changing information.
- Shadows should resemble offset paper layers rather than diffuse floating
  glass panels.

### Icons, Sound, And Motion

- Prefer simple stamped, printed, or utilitarian line icons.
- Avoid food emoji as core interface iconography.
- Use a short cue at `GO`, a quiet tick, stamp, or bell for new ingredients,
  and distinct correct, incorrect, and round-end sounds.
- New ingredients may enter with a quick ticket-stamp motion; success may use a
  check stamp in Herb.
- Effects are brief and optional, with a mute control. They never determine or
  delay authoritative game timing.

### Visual Guardrails

- Do not treat the visual direction as a requirement to imitate point-of-sale
  software.
- Do not sacrifice scan speed for decorative restaurant realism.
- Keep the Restaurant Ticket motif consistent rather than mixing in unrelated
  visual styles.
