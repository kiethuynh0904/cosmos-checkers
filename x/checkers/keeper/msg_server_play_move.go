package keeper

import (
	"context"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/kiethuynh0904/checkers/x/checkers/rules"
	"github.com/kiethuynh0904/checkers/x/checkers/types"
)

func (k msgServer) PlayMove(goCtx context.Context, msg *types.MsgPlayMove) (*types.MsgPlayMoveResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	//Fetch the stored game information
	storedGame, found := k.Keeper.GetStoredGame(ctx, msg.IdValue)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrGameNotFound, "game not found %s", msg.IdValue)
	}

	//Is the player legitimate?
	var player rules.Player
	if strings.Compare(storedGame.Red, msg.Creator) == 0 {
		player = rules.RED_PLAYER
	} else if strings.Compare(storedGame.Black, msg.Creator) == 0 {
		player = rules.BLACK_PLAYER
	} else {
		return nil, types.ErrCreatorNotPlayer
	}

	// Instantiate the board to implement the rules:
	game, err := storedGame.ParseGame()
	if err != nil {
		panic(err.Error())
	}

	//Is it the player's turn?
	if !game.TurnIs(player) {
		return nil, types.ErrNotPlayerTurn
	}

	//Properly conduct the move
	captured, moveErr := game.Move(
		rules.Pos{
			X: int(msg.FromX),
			Y: int(msg.FromY),
		},
		rules.Pos{
			X: int(msg.ToX),
			Y: int(msg.ToY),
		},
	)
	if moveErr != nil {
		return nil, sdkerrors.Wrapf(moveErr, types.ErrWrongMove.Error())
	}

	// Prepare the updated board to be stored and store the information
	storedGame.Game = game.String()
	storedGame.Turn = game.Turn.Color
	k.Keeper.SetStoredGame(ctx, storedGame)

	//Return relevant information regarding the move's result
	return &types.MsgPlayMoveResponse{
		IdValue:   msg.IdValue,
		CapturedX: int64(captured.X),
		CapturedY: int64(captured.Y),
		Winner:    game.Winner().Color,
	}, nil
}
