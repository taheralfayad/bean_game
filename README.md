# The Bean Game

this is an implementation of my favorite board game, Bohnanza

## Testing game events

Game events are streamed over Server-Sent Events at `GET /gameevents/:topic`.
Open a subscriber in one terminal:

```bash
curl -N http://localhost:8080/gameevents/game:12
```

Then drive the game from a second terminal:

```bash
curl -X POST http://localhost:8080/games
curl -X POST http://localhost:8080/games/12/join
curl -X POST http://localhost:8080/games/12/start
```

Use `curl -N` (or a client that streams responses); clients that buffer the
full body will appear to hang.
