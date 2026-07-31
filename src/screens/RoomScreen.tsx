type RoomScreenProps = {
  match: import("../types").MatchSnapshot;
  onLeave: () => void;
};

export function RoomScreen({ match, onLeave }: RoomScreenProps) {
  const inviteUrl = `${window.location.origin}/room/${match.roomCode}`;

  return (
    <main className="centered-screen">
      <section className="room-ticket">
        <span className="status-label">Private table</span>
        <h1>Your table is reserved</h1>
        <p>Send the invite to the person you want to challenge.</p>
        <div className="invite-link">
          <span>{inviteUrl}</span>
          <button onClick={() => navigator.clipboard?.writeText(inviteUrl)}>Copy</button>
        </div>
        <div className="room-seats">
          <div><small>Host</small><strong>{match.player.name}</strong><span>Ready</span></div>
          <div><small>Guest</small><strong>Waiting...</strong><span>Empty seat</span></div>
        </div>
        <p className="answer-hint">The match starts when your guest joins.</p>
        <button className="text-button" onClick={onLeave}>Leave room</button>
      </section>
    </main>
  );
}
