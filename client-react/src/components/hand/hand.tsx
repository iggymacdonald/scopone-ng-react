import { Button } from "@material-ui/core";
import React, { FC, useContext, useEffect, useState } from "react";
import { combineLatest, merge } from "rxjs";
import { tap } from "rxjs/operators";
import { ServerContext } from "../../context/server-context";
import { Card, TypeValues } from "../../rx-services/scopone-rx-service/card";
import { HandState, Team } from "../../rx-services/scopone-rx-service/messages";
import { Cards } from "../cards/cards";
import { Table } from "../table/table";

// we define a type for the state so that we can issue a single call to the update state function and
// avoid so multiple execution of the render function
// https://stackoverflow.com/questions/53574614/multiple-calls-to-state-updater-from-usestate-in-component-causes-multiple-re-re
type HandReactState = {
  showStartButton: boolean;
  teams?: [Team, Team];
  playerCards: Card[];
  table?: Card[];
  ourScope: Card[];
  theirScope: Card[];
  currentPlayerName?: string;
  enablePlay: boolean;
};

export const Hand: FC = () => {
  const server = useContext(ServerContext);

  const [handReactState, setHandReactState] = useState<HandReactState>({
    playerCards: [],
    ourScope: [],
    theirScope: [],
    showStartButton: false,
    enablePlay: false,
  });

  useEffect(() => {
    console.log("=======>>>>>>>>>>>>  Use Effect run in Hand");

    // myCurrentGame$ Observable sets teams and showSartButton state as a side effect
    // when the updated info about my current game is notified on the server stream
    const myCurrentGame$ = server.myCurrentOpenGame_ShareReplay$.pipe(
      tap((game) => {
        const teams = game.teams;
        // decide whether to show or not the Start Game button
        const gameWith4PlayersAndNoHand =
          Object.keys(game.players).length === 4 && game.hands.length === 0;
        const lastHandClosed = game.hands
          ? game.hands.length > 0
            ? game.hands[game.hands.length - 1].state === HandState.closed
            : false
          : false;
        const showStartButton = gameWith4PlayersAndNoHand || lastHandClosed;
        setHandReactState((prevState) => ({
          ...prevState,
          teams,
          showStartButton,
        }));
      })
    );

    // handView$ Observable sets teams and showSartButton state as a side effect
    // when the updated info about my current OBSERVED game is notified on the server stream
    const myObservedGame$ = server.myCurrentObservedGame_ShareReplay$.pipe(
      tap((game) => {
        const teams = game.teams;
        setHandReactState((prevState) => ({ ...prevState, teams }));
      })
    );

    // myCurrentGame$ Observable sets cards, scope and currentPlayerName as a side effect
    // when new hand views are notified on the server stream
    const handView$ = server.handView_ShareReplay$.pipe(
      tap((hv) => {
        const pCards = hv.playerCards?.sort(
          (a, b) => TypeValues.get(b.type) - TypeValues.get(a.type)
        );
        const newState: Partial<HandReactState> = {
          playerCards: pCards,
          table: hv.table,
          ourScope: hv.ourScope,
          theirScope: hv.theirScope,
          currentPlayerName: hv.currentPlayerName,
        };
        setHandReactState((prevState) => ({ ...prevState, ...newState }));
      })
    );

    const enablePlay$ = combineLatest([
      server.isMyTurnToPlay$,
      server.myCurrentOpenGameWithAll4PlayersIn_ShareReplay$,
    ]).pipe(
      tap(([isMyTurn, all4PlayersIn]) => {
        const enablePlay = isMyTurn && all4PlayersIn;
        setHandReactState((prevState) => ({ ...prevState, enablePlay }));
      })
    );

    const subscription = merge(
      myCurrentGame$,
      myObservedGame$,
      handView$,
      enablePlay$
    ).subscribe();

    return () => {
      console.log("Unsubscribe Hand subscription");
      subscription.unsubscribe();
    };
  }, [server]);

  const start = () => {
    server.newHand();
  };

  const play = (card: Card) => {
    //this.canSendCardToServer = false;
    const cardsTakeable = server.cardsTakeable(card, handReactState.table);
    if (cardsTakeable.length > 1) {
      // const dialogRef = this.dialog.open(CardsPickerDialogueComponent, {
      //   width: '1250px',
      //   height: '600px',
      //   data: { cards: cardsTakeable },
      // });
      // dialogRef.afterClosed().subscribe((cardsToTake) => {
      //   console.log('The player has chosen', cardsToTake);
      //   this.scoponeService.playCardForPlayer(
      //     this.scoponeService.playerName,
      //     card,
      //     cardsToTake
      //   );
      // });
      throw new Error("multiple choice not yet implemented");
    } else {
      server.playCard(card, handReactState.table);
    }
  };

  return (
    <>
      {handReactState.teams && (
        <Table
          teams={handReactState.teams}
          currentPlayerName={handReactState.currentPlayerName}
          cards={handReactState.table}
        ></Table>
      )}
      {handReactState.playerCards.length > 0 && (
        <Cards
          cards={handReactState.playerCards}
          name="My cards"
          initialLayout="spread"
          cardClickHandler={play}
          enabled={handReactState.enablePlay}
        ></Cards>
      )}
      {handReactState.ourScope.length > 0 && (
        <Cards cards={handReactState.ourScope} name="Our Scope"></Cards>
      )}
      {handReactState.theirScope.length > 0 && (
        <Cards cards={handReactState.theirScope} name="Their Scope"></Cards>
      )}
      {handReactState.showStartButton && (
        <Button size="small" onClick={start}>
          Start
        </Button>
      )}
    </>
  );
};
