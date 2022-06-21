package tarabish

import (
	"fmt"

	"go-tarabish/src/game-logic/deck"
	"go-tarabish/src/game-logic/player"
	"go-tarabish/src/game-logic/team"
)

// State is the state of a gamew
type State string

// GameState possible values
const (
	GameCreated   State = "created"      // created but with no players yet"
	TeamsForming  State = "teamsForming" // it has some players but not all
	GameOpen      State = "open"
	GameSuspended State = "suspended"
	GameClosed    State = "closed"
)

// Game represents a match of Scopone
type Game struct {
	Name      string                    `json:"name"`
	Hands     []*Hand                   `json:"hands"`
	Teams     []*team.Team              `json:"teams"`
	Players   map[string]*player.Player `json:"players"`
	Observers map[string]*player.Player `json:"observers"`
	Score     map[string]int            `json:"score"`
	State     State                     `json:"state"`
	ClosedBy  string                    `json:"closedBy"`
	History   []*HandHistory            `json:"-"`
}

// NewGame game
func NewGame() *Game {
	fmt.Println("A new Game is created")
	g := Game{}
	g.Teams = make([]*team.Team, 2)
	g.Teams[0] = team.New()
	g.Teams[1] = team.New()
	g.Players = make(map[string]*player.Player)
	g.Observers = make(map[string]*player.Player)
	g.Score = make(map[string]int)
	g.Hands = make([]*Hand, 0)
	g.State = GameCreated
	return &g
}

// Suspend suspends the game
func (game *Game) Suspend() {
	game.State = GameSuspended
}

// Close the game and sets all other players as not playing
// this means that if just ONE player leaves the game, all other players leave it
func (game *Game) Close(playerClosing string) {
	if game.State == GameClosed {
		return
	}
	for pK := range game.Players {
		p := game.Players[pK]
		p.Status = player.PlayerNotPlaying
	}
	game.State = GameClosed
	game.ClosedBy = playerClosing
}

type handState string

const (
	// HandActive the hand is active
	HandActive handState = "active"
	// HandClosed the hand is closed
	HandClosed handState = "closed"
)

// Hand is an hand of a Tarabish Match
type Hand struct {
	Deck          deck.Deck `json:"-"`
	State         handState `json:"state"`
	Winner        team.Team
	CurrentDealer *player.Player
	FirstPlayer   *player.Player
	CurrentPlayer *player.Player
	Table         []deck.Card          `json:"-"`
	Score         map[string]TeamScore `json:"-"`
	History       HandHistory          `json:"-"`
}

// HandCardPlay represents a single card played by a player with the cards it took
type HandCardPlay struct {
	Player       string                 `json:"player"`
	Table        []deck.Card            `json:"table"`
	CardPlayed   deck.Card              `json:"cardPlayed"`
	CardsTaken   []deck.Card            `json:"cardsTaken"`
	PlayersDecks map[string][]deck.Card `json:"playersDecks"`
}

// HandHistory contains the hystory of the hand
type HandHistory struct {
	PlayerDecks      map[string][]deck.Card `json:"playerDecks"`
	CardPlaySequence []HandCardPlay         `json:"cardPlaySequence"`
}

// AddPlayer adds a player to a game and to one of the 2 teams
func (game *Game) AddPlayer(p *player.Player) error {
	if len(game.Players) == 4 {
		var playerNames string
		for pName := range game.Players {
			playerNames = playerNames + " " + pName
		}
		return fmt.Errorf("Game has already 4 Players: %v", playerNames)
	}
	// the same player can not be added twice to the same game
	_, pFound := game.Players[p.Name]
	if pFound {
		return fmt.Errorf("Player %v is already present in game %v", p.Name, game.Name)
	}
	// the player fills the first slot free in the teams - this allows a player to reenter a game at his place
	switch noOfPlayer := len(game.Players); noOfPlayer {
	case 0:
		game.Teams[0].Players[0] = p
	case 1:
		game.Teams[0].Players[1] = p
	case 2:
		game.Teams[1].Players[0] = p
	case 3:
		game.Teams[1].Players[1] = p
	}
	game.Players[p.Name] = p
	p.Status = player.PlayerPlaying
	game.CalculateState()
	return nil
}

// AddObserver adds an Observer to a game
func (game *Game) AddObserver(p *player.Player) error {
	// the same observer can not be added twice to the same game
	_, oFound := game.Observers[p.Name]
	if oFound {
		return fmt.Errorf("%v is already observing game %v", p.Name, game.Name)
	}
	game.Observers[p.Name] = p
	p.Status = player.PlayerObserving
	return nil
}

// CalculateState calculates the sate of the game
func (game *Game) CalculateState() {
	if len(game.Players) == 0 {
		game.State = GameCreated
		return
	}
	if game.State == GameClosed {
		return
	}
	for kP := range game.Players {
		p := game.Players[kP]
		if p.Status == player.PlayerLeftOsteria {
			game.State = GameSuspended
			return
		}
		if p.Status != player.PlayerPlaying && p.Status != player.PlayerLookingAtHandResult {
			msg := fmt.Sprintf(`Player %v in game %v has state "%v" which is never expected to happen 
			since player in a game should either be playing or be suspended`, p.Name, game.Name, p.Status)
			panic(msg)
		}
	}
	if len(game.Players) == 4 {
		game.State = GameOpen
		return
	}
	game.State = TeamsForming
}

// HandPlayerView is the data set that a Player can see of a running hand
type HandPlayerView struct {
	ID                    string      `json:"id"`
	GameName              string      `json:"gameName"`
	PlayerCards           []deck.Card `json:"playerCards"`
	Table                 []deck.Card `json:"table"`
	OurScope              []deck.Card `json:"ourScope"`   // Scope of the player's team
	TheirScope            []deck.Card `json:"theirScope"` // Scope of the other team
	OurScorecard          ScoreCard   `json:"ourScorecard"`
	TheirScorecard        ScoreCard   `json:"theirScorecard"`
	Status                handState   `json:"status"`
	FirstPlayerName       string      `json:"firstPlayerName"`
	CurrentPlayerName     string      `json:"currentPlayerName"`
	OurCurrentGameScore   int         `json:"ourCurrentGameScore"`
	TheirCurrentGameScore int         `json:"theirCurrentGameScore"`
	OurFinalHandScore     int         `json:"ourFinalScore"`
	TheirFinalHandScore   int         `json:"theirFinalScore"`
	History               HandHistory `json:"history,omitempty"`
}

// ScoreCard organizes the cards to facilitate calculating the score of a Team
type ScoreCard struct {
	// The point is won by whichever team takes the K and Q of trump by the same player, known as the 'the bells' .
	Bella bool `json:"bella"`
	// Last is awarded to the team that takes the last hand
	Last bool `json:"last"`
	// 50 points are awarded to a team that has 4 consecutive cards of the same suit - it is possible to have up to 2 fifties
	Fifties []deck.Card `json:"fifties"`
	// 50 points are awarded to a team that has 4 consecutive cards of the same suit - it is possible to have up to 2 fifties
	Twenties []deck.Card `json:"twenties"`
	// Cards are the slice of cards taken by one of the teams
	Cards []deck.Card `json:"cards"`

	// The point is won by whichever team takes more cards of the coins suit (or diamonds if you are using international cards). If they split 5-5 the point is not awarded.
	Denari []deck.Card `json:"denari"`
	// The point is won by whichever team takes the 7 of coins (diamonds), known as the 'sette bello' (beautiful seven).
	Settebello bool `json:"settebello"`
	//The point is won by the team with the best prime. In practice this is usually the team with more sevens,
	//but the actual rule is as follows. A prime consists of one card of each suit, and the cards have special
	//point values for this purpose, as shown in the table. The value of the prime is got by adding up the
	//values of its cards and whichever team can construct the more valuable prime wins the point
	PrimieraSuits map[string][]deck.Card `json:"primiera"`
	// The Cards. The point is won by whichever team takes the majority of the cards. If they split 20-20 the point is not awarded
	Carte []deck.Card `json:"carte"`
	// sweeps - In addition to the points mentioned above, you also win a point for each sweep (Italian scopa).
	// You score a sweep when you play a card which captures the all table cards, leaving the table empty.
	// Traditionally, the capturing card is placed face up in the trick-pile of the capturing side,
	// so that the number of sweeps made by each side can easily be seen when the scoring is done at the end of the play.
	Scope []deck.Card `json:"scope"`
	//Some play that a team that captures the ace, two and three of coins scores a number of points equal
	// to the highest coin card they capture in unbroken sequence with these - for example if they took
	// the A-2-3-4-5-6 of coins they would score 6 points (in addition to the point for coins).
	// This bonus is called Napola or Napoli. A team that captures all ten cards of the coin suit wins
	// the game outright. This is called Napoleone or Napolone or Cappotto
	Napoli []deck.Card `json:"napoli"`
}

// TeamScore is a data struct containg info related to the score of a team in one hand
type TeamScore struct {
	ScoreCard     ScoreCard
	BonusScore    int
	PrimieraScore int
	Score         int
}
