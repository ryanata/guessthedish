# Guess the Dish

Provisional product and interaction specification, version 0.1.

This document captures the current direction for the prototype. It is a
starting point for critique, playtesting, and revision rather than a final
contract. Items marked **Prototype default** are intentionally easy to change.

## Product Summary

Guess the Dish is a fast, one-versus-one web game. A dish's ingredients are
revealed one at a time, and players race to identify the dish by selecting an
answer from a fixed culinary catalog.

The central decision is whether to guess early with incomplete information or
wait for a more distinctive ingredient and risk losing the race.

## Product Principles

- **Understandable immediately.** A new player should understand the game by
  watching one round.
- **Knowledge over loopholes.** Winning should come from recognizing a dish,
  not exploiting fuzzy free-text adjudication or spamming answers.
- **Fast recovery.** Wrong guesses matter without taking a player out of the
  round for long.
- **Symmetric competition.** Both players see the same clues in the same order
  at the same server-controlled times.
- **Short sessions.** A complete match should fit comfortably into a casual
  break.
- **No empty multiplayer.** Quick Play can be filled by a bot so the prototype
  remains playable with a small audience.

## Terminology

- **Dish:** A canonical answer in the answer catalog, such as `Spaghetti
  Bolognese`.
- **Alias:** An accepted searchable name that resolves to a canonical dish,
  such as `spag bol`.
- **Puzzle:** A dish plus its curated, ordered ingredient clues.
- **Round:** One puzzle played until a correct answer or timeout.
- **Match:** A sequence of rounds until one player wins three rounds.
- **Room:** A temporary match lobby reached through an invite URL.

## Match Format

- Matches are one versus one.
- The first player to win three rounds wins the match.
- A round win awards one point.
- An unsolved round awards no point to either player.
- Unsolved rounds do not advance either player's score; the match continues
  with a new puzzle.
- A puzzle must not repeat within a match.
- The first correct answer accepted by the authoritative game server ends the
  round immediately.
- There is no grace window or tied round in the prototype.

### Prototype Defaults

- Three-second countdown before each round.
- Four-second result state between rounds.
- A match result remains visible until the player chooses to play again or
  leave.

## Round And Ingredient Timing

Puzzles may contain different numbers of displayed ingredient clues. They are
not forced into a seven-clue template.

- Minimum puzzle size: four ingredients.
- **Prototype default:** maximum puzzle size: twelve ingredients.
- Every revealed ingredient remains visible for the rest of the round.
- The first ingredient appears at `GO`.
- Ingredients are curated from less identifying to more identifying.
- Both players receive exactly the same reveal order and schedule.
- Each clue shows an ingredient label only, not its quantity.
- Preparation may be part of the label only when it changes the ingredient's
  identity in a meaningful way, for example `roasted red pepper`.

Let:

- `N` be the number of ingredients in the puzzle.
- `K = ceil(N * 0.25)` be the number of early clues.

Timing:

- After each of the first `K` reveals, wait two seconds before revealing the
  next ingredient.
- After each remaining reveal, wait three seconds before revealing the next
  ingredient.
- After the final ingredient, leave a final three-second answer window.
- If nobody answers correctly by the end of that window, reveal the dish and
  award no point.

The final clue's three-second answer window is included in the maximum round
times below.

| Ingredient count | Early clues (`K`) | Maximum round length |
| ---: | ---: | ---: |
| 4 | 1 | 11 seconds |
| 5 | 2 | 13 seconds |
| 7 | 2 | 19 seconds |
| 8 | 2 | 22 seconds |
| 10 | 3 | 27 seconds |
| 12 | 3 | 33 seconds |

The formula and maximum should be playtested. Long recipes may eventually use
a curated subset rather than every source ingredient.

## Answering

Players do not submit arbitrary prose. They search and select from a static
catalog of valid dishes.

### Search And Selection

- Search begins after two typed characters.
- Show at most five results.
- Search covers canonical dish names and aliases.
- Search never considers ingredient lists or semantic similarity.
- Results always display the canonical dish name.
- A selected result is submitted immediately.
- Pressing `Enter` submits the highlighted result, or the first result when
  none is highlighted.
- Arrow keys move through results.
- Freeform text that does not resolve to a catalog entry cannot be submitted.
- Exact aliases rank ahead of partial matches.
- Conservative typo matching is permitted, but it must still resolve to a
  static catalog entry.

Example:

```text
Query: bolognese

Spaghetti Bolognese
Ragù alla Bolognese
```

### Wrong Answers

- An incorrect submission locks that player's answer control for 750 ms.
- The lock does not pause ingredient reveals.
- The rejected dish cannot be submitted again during that round.
- Previously rejected results may remain visible but should be visibly marked
  and disabled.
- The opponent's guesses remain hidden during the round.

### Correct Answers

- The server, not the browser, determines which correct answer arrived first.
- The first accepted correct answer ends the round for both players.
- The result state identifies the winner and the ingredient number visible
  when the answer was accepted.

## Puzzle Content Model

Each puzzle is curated rather than generated directly from an arbitrary source
recipe.

Required fields:

- Stable puzzle identifier.
- Canonical dish identifier and display name.
- Searchable aliases.
- At least four ordered ingredient clues.
- Internal difficulty rating.
- Content status such as `draft`, `reviewed`, or `retired`.

Optional fields:

- Cuisine or region.
- Dish photograph and attribution.
- Short post-round context.
- Source references for editorial review.

### Content Guidelines

- A clue set should converge on a conventional named dish rather than only a
  broad category.
- Generic ingredients should normally appear before distinctive ingredients.
- A highly revealing eponymous or specialty ingredient should not appear first
  unless the puzzle is intentionally easy.
- Salt, water, neutral oil, and similar pantry defaults may be omitted when
  they add no useful signal.
- Variants should be separate answers only when their clue sets meaningfully
  distinguish them.
- Aliases are equivalences, not broad descriptions. `Spag bol` can resolve to
  `Spaghetti Bolognese`; `pasta with meat sauce` should not.
- Frequently disputed or poorly performing puzzles should be revised or
  retired.

## Play Modes

### Quick Play

1. The player supplies or accepts a generated guest name.
2. The player selects **Quick Play**.
3. The system searches for a compatible waiting player.
4. **Prototype default:** if no player is available after three seconds, a bot
   fills the opposing seat.
5. The match begins when both seats are ready.

Bots should use the revealed information, plausible reaction delays, and a
configurable skill profile. They should not merely know the answer and wait a
fixed random duration. Whether bots are visually identified is an open product
question.

### Invite Room

1. The host selects **Create Room**.
2. The system creates a temporary room and shareable invite URL.
3. The host can copy the URL while waiting.
4. The first valid guest to open or join through the URL claims the second
   player seat.
5. The match starts automatically after both players are connected and ready.
6. A third visitor sees a clear **Room full** state.
7. Private rooms do not receive automatic bot fill.
8. After a match, both players may opt into a rematch in the same room.

No account, friend list, or social graph is required. Room URLs should be
unguessable and temporary.

### Reconnection Defaults

- A brief disconnect does not immediately forfeit the match.
- **Prototype default:** reserve the player's seat for 15 seconds.
- The authoritative round timeline continues during disconnection.
- If the player returns in time, the client catches up to the current state.
- If the player does not return, the remaining player wins the match by
  forfeit.

## Screen Specifications

### Home

Purpose: explain the premise and get the player into a match with minimal
friction.

Required elements:

- Guess the Dish wordmark.
- One-sentence explanation.
- Editable guest display name.
- Primary **Quick Play** action.
- Secondary **Create Room** action.
- Compact how-to-play demonstration.

No account is required.

### Quick Play Matchmaking

- Visible searching state.
- Elapsed search feedback or lightweight animation.
- Cancel action.
- Transition directly into the match countdown when an opponent is found.

### Invite Room Lobby

- Room identifier or concise waiting label.
- Host and guest player slots.
- Shareable URL.
- **Copy Invite Link** action with confirmation feedback.
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
- Mute control.

The active game must not require page scrolling at supported viewport sizes.

### Round Result

- Canonical dish name.
- Winning player, or **No one got it**.
- Ingredient number on which the winning answer was accepted.
- Full ordered clue sequence.
- Updated match score.
- Dish image when available and properly licensed.
- Brief automatic transition to the next round.

### Match Result

- Winner and final score.
- Round-by-round summary.
- **Play Again** for Quick Play.
- Rematch readiness for invite rooms.
- Leave action.

## Responsive And Accessible Interaction

- Mobile-first layout with an expanded desktop composition.
- The game board must fit without scrolling during an active round.
- On mobile, autocomplete opens upward so results remain visible above the
  software keyboard.
- Nonessential decoration collapses while the keyboard is open.
- Touch targets are at least 44 by 44 CSS pixels.
- The complete game is keyboard operable.
- Focus states are clearly visible.
- Status never relies on color alone.
- Announce new clues, lockouts, round results, and connection changes to
  assistive technology without repeatedly stealing focus.
- Respect `prefers-reduced-motion` and provide a mute control.
- Maintain readable contrast across all selected visual themes.

## Sound And Motion

Sound and animation should reinforce state changes without delaying play.

Prototype candidates:

- Short cue at `GO`.
- Quiet tick, stamp, or service-bell cue for each new ingredient.
- Distinct correct, incorrect, and round-end sounds.
- A quick entrance animation for a newly revealed clue.

All effects must be optional. Functional timing cannot depend on animation
completion.

## Visual Direction Exploration

The final visual system is intentionally undecided. `designs.html` is the
comparison artifact for evaluating six initial directions against the same
game scaffold:

1. Restaurant Ticket
2. Bistro Chalkboard
3. Retro Diner
4. Stainless Kitchen
5. Midnight Kitchen
6. Editorial Cookbook

Each direction includes its palette, typography notes, game controls, answer
results, ingredient states, and invite-room treatment. The comparison page is
not production application code.

## Prototype Scope

### Included

- Guest identity.
- One-versus-one Quick Play.
- Bot fallback for Quick Play.
- Shareable invite rooms.
- First-to-three matches.
- Variable-length, server-timed ingredient reveals.
- Static answer catalog with aliases and conservative typo matching.
- Wrong-answer lockouts.
- Curated initial puzzle set.
- Synchronized round and result states.
- Responsive desktop and mobile gameplay.
- Basic reconnect handling.

### Excluded

- User accounts and persistent profiles.
- Rankings, matchmaking ratings, and leaderboards.
- Friend lists or a social graph.
- Chat and direct messaging.
- User-authored recipes or puzzles.
- Daily challenges, achievements, and progression systems.
- Monetization.
- General recipe browsing outside the post-round context.
- Native mobile applications.

## Prototype Success Questions

The prototype should answer these questions before expanding scope:

- Do players understand the game without instruction?
- Is there meaningful tension between guessing early and waiting?
- Does autocomplete feel fast, fair, and resistant to answer spam?
- Is a 750 ms penalty perceptible enough without becoming frustrating?
- Does the variable timing make short and long recipes feel equally paced?
- How often do players dispute a puzzle or expected alias?
- Is first-to-three the right session length?
- How frequently does a human choose an answer before the final clue?
- Do invite links produce a smooth two-person start and rematch loop?
- Can bots fill empty matches without making play feel mechanical?

## Open Decisions

- Final visual direction or combination of directions.
- Whether Quick Play bots are explicitly labeled.
- Initial puzzle count and cuisine distribution.
- Exact typo-matching tolerance.
- Whether autocomplete should expose already rejected answers or remove them.
- Whether puzzle difficulty affects matchmaking or remains mixed.
- Whether the twelve-ingredient maximum is correct.
- Whether long source recipes should be represented by every meaningful
  ingredient or a curated subset.
- Room expiration duration.
- Host behavior when the invited guest disconnects before a match.
- Photo sourcing and licensing policy.
- Content moderation and editorial workflow after the prototype.
