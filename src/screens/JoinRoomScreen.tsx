type JoinRoomScreenProps = {
  guestName: string;
  error?: string;
  onGuestNameChange: (name: string) => void;
  onJoin: () => void;
  onLeave: () => void;
};

export function JoinRoomScreen({ guestName, error, onGuestNameChange, onJoin, onLeave }: JoinRoomScreenProps) {
  return (
    <main className="centered-screen">
      <section className="room-ticket">
        <span className="status-label">Private table</span>
        <h1>You’re invited</h1>
        <p>Choose a name and take the open seat.</p>
        <div className="room-join-controls">
          <label htmlFor="room-guest-name">Guest name</label>
          <input
            id="room-guest-name"
            maxLength={40}
            placeholder="Enter your name"
            value={guestName}
            onChange={(event) => onGuestNameChange(event.target.value)}
          />
          <button className="primary-button" onClick={onJoin}>Join table</button>
          {error && <p className="form-error" role="alert">{error}</p>}
        </div>
        <button className="text-button" onClick={onLeave}>Back to kitchen</button>
      </section>
    </main>
  );
}
