import { ChefIllustration } from "../components/ChefIllustration";

type HomeScreenProps = {
  guestName: string;
  onGuestNameChange: (name: string) => void;
  onQuickPlay: () => void;
  onCreateRoom: () => void;
  error?: string;
};

function JoinControls({ guestName, onGuestNameChange, onQuickPlay, onCreateRoom, error }: HomeScreenProps) {
  return (
    <div className="duel-controls">
      <label htmlFor="guest-name">Guest name</label>
      <input
        id="guest-name"
        maxLength={20}
        placeholder="Enter your name"
        value={guestName}
        onChange={(event) => onGuestNameChange(event.target.value)}
      />
      <button className="primary-button" onClick={onQuickPlay}>Quick Play</button>
      <button className="secondary-button" onClick={onCreateRoom}>Create Room</button>
      {error && <p className="form-error" role="alert">Backend unavailable: {error}</p>}
    </div>
  );
}

export function HomeScreen(props: HomeScreenProps) {
  return (
    <main className="home-screen">
      <section className="clean-duel-stage">
        <div className="clean-duel-copy">
          <h1>Guess <em>the Dish</em></h1>
          <p>Ingredients land one by one. Name the dish before your opponent does.</p>
        </div>

        <div className="clean-duel-chefs">
          <ChefIllustration color="red" />
          <span className="clean-vs">VS</span>
          <ChefIllustration color="green" facing="left" />
        </div>

        <JoinControls {...props} />
      </section>
    </main>
  );
}
