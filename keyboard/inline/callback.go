package inline

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/jkevinp/tgui/uibot"
)

func (kb *Keyboard) callbackAnswer(ctx *uibot.Context, callbackQuery *models.CallbackQuery) {
	ok, err := ctx.BotInstance.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: callbackQuery.ID,
	})
	if err != nil {
		kb.onError(err)
		return
	}
	if !ok {
		kb.onError(fmt.Errorf("callback answer failed"))
	}
}

func (kb *Keyboard) callback(ctx *uibot.Context, update *models.Update) {
	if kb.deleteAfterClick {

		b := ctx.BotInstance
		b.UnregisterHandler(kb.callbackHandlerID)

		_, errDelete := b.DeleteMessage(ctx, &bot.DeleteMessageParams{
			ChatID:    update.CallbackQuery.Message.Message.Chat.ID,
			MessageID: update.CallbackQuery.Message.Message.ID,
		})
		if errDelete != nil {
			kb.onError(fmt.Errorf("error delete message in callback, %w", errDelete))
		}
	}

	btnNum, errBtnNum := strconv.Atoi(strings.TrimPrefix(update.CallbackQuery.Data, kb.prefix))
	if errBtnNum != nil {
		kb.onError(fmt.Errorf("wrong callback data btnNum, %s", update.CallbackQuery.Data))
		kb.callbackAnswer(ctx, update.CallbackQuery)
		return
	}

	if len(kb.handlers) <= btnNum {
		kb.onError(fmt.Errorf("wrong callback data, %s", update.CallbackQuery.Data))
		kb.callbackAnswer(ctx, update.CallbackQuery)
		return
	}

	if kb.handlers[btnNum].Handler != nil {
		kb.handlers[btnNum].Handler(ctx, update.CallbackQuery.Message, kb.handlers[btnNum].data)
	}

	kb.callbackAnswer(ctx, update.CallbackQuery)
}
