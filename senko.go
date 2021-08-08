package Senko

import (
	"encoding/json"
	"fmt"
	tgbotapi "github.com/4Kaze/telegram-bot-api/v5"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
)

const (
	CHAT_TYPE_PRIVATE    = "private"
	CHAT_TYPE_GROUP      = "group"
	CHAT_TYPE_SUPERGROUP = "supergroup"
	MAX_NAME_CHARACTERS  = 20
	EMOJI_REGEX          = "[\U00010000-\U0010ffff]"
	SYMBOL_REGEX         = "[@#$-/:-?{-~!\"^_`\\[\\]]"
	COMMAND              = "drawtext=fontfile=%s/%s:text='Welcome to the group\\! %s-kun\\!':bordercolor=black:borderw=1:fontcolor=white:fontsize=25:x=(w-text_w)/2:y=h-th-20:enable='gte(t,1.5)'"
	INPUT_FILE           = "%s/greeting.mp4"
	GCP_DIR              = "./serverless_function_source_code"

	START_REPLY = `Wewcome! OwO
Senko-san onwy wowks on gwouwps. Juwst add me to uw gwouwp and i'ww gweat aww newcomews.
If u want me to give u a gweeting gif with cuwstom name, uwse command /genewate [name]. That command onwy wowks hewe and not in gwouwps UwU
Use command /wepo to get a wink to my souwwce code on GitHuwb.`
	REPO_URL                   = "https://github.com/4Kaze/telegram-senko-bot"
	GENERATE_USAGE             = "Usage: /genewate [name], whewe [name] is uw name OwO"
	GENERATION_STARTED_MESSAGE = "Gotcha! OwO howd on, it wiww take a dozen seconds"
	START_COMMAND              = "/start"
	REPO_COMMAND               = "/wepo"
	GENERATE_COMMAND           = "/genewate"
)

var (
	once        sync.Once
	emojiRegex  *regexp.Regexp
	symbolRegex *regexp.Regexp
	bot         *tgbotapi.BotAPI
	resourceDir = GCP_DIR
	fontFile    = "NotoSansCJKjp-Black.otf"
)

func init() {
	var err error
	emojiRegex, err = regexp.Compile(EMOJI_REGEX)
	if err != nil {
		log.Fatalln(err)
	}
	symbolRegex, err = regexp.Compile(SYMBOL_REGEX)
	if err != nil {
		log.Fatalln(err)
	}
}

func initBot() {
	var err error
	token := os.Getenv("TOKEN")
	bot, err = tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal(err)
	}
}

func HandleRequest(w http.ResponseWriter, r *http.Request) {
	once.Do(initBot)
	var update tgbotapi.Update
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		fmt.Println("Could not decode request")
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	if err := handleUpdate(update); err != nil {
		fmt.Println(err)
		_, _ = w.Write([]byte(err.Error()))
	}
}

func handleUpdate(update tgbotapi.Update) error {
	if !isValidUpdate(update) {
		return nil
	}
	if isGroupGreetingUpdate(update) {
		return handleGroupGreeting(update)
	} else if isPrivateChatCommandUpdate(update) {
		return handlePrivateChatCommand(update)
	}
	return nil
}

func isValidUpdate(update tgbotapi.Update) bool {
	return update.Message != nil && update.Message.Chat != nil
}

func isGroupGreetingUpdate(update tgbotapi.Update) bool {
	return isGroupUpdate(update) && update.Message.NewChatMembers != nil && len(*update.Message.NewChatMembers) != 0
}

func isGroupUpdate(update tgbotapi.Update) bool {
	return update.Message.Chat.Type == CHAT_TYPE_GROUP || update.Message.Chat.Type == CHAT_TYPE_SUPERGROUP
}

func isPrivateChatCommandUpdate(update tgbotapi.Update) bool {
	return update.Message.Chat.Type == CHAT_TYPE_PRIVATE && strings.HasPrefix(update.Message.Text, "/")
}

func handleGroupGreeting(update tgbotapi.Update) error {
	name := (*update.Message.NewChatMembers)[0].FirstName
	return generateAndSend(name, update.Message.Chat.ID, update.Message.MessageID)
}

func handlePrivateChatCommand(update tgbotapi.Update) error {
	command := strings.Fields(update.Message.Text)[0]
	switch command {
	case START_COMMAND:
		return sendTextMessage(update.Message.Chat.ID, START_REPLY)
	case REPO_COMMAND:
		return sendTextMessage(update.Message.Chat.ID, REPO_URL)
	case GENERATE_COMMAND:
		return handleGenerateCommand(update)
	}
	return nil
}

func handleGenerateCommand(update tgbotapi.Update) error {
	trimmedName := strings.TrimSpace(strings.TrimPrefix(update.Message.Text, GENERATE_COMMAND))
	if len(trimmedName) == 0 {
		return sendTextMessage(update.Message.Chat.ID, GENERATE_USAGE)
	}
	if err := sendTextMessage(update.Message.Chat.ID, GENERATION_STARTED_MESSAGE); err != nil {
		return err
	}
	return generateAndSend(trimmedName, update.Message.Chat.ID, 0)
}

// replyToMessageId - 0 when not a reply
func generateAndSend(name string, chatId int64, replyToMessageId int) error {
	file, err := ioutil.TempFile("", "*.mp4")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	err = generateGif(stripName(name), file.Name())
	if err != nil {
		return err
	}
	return sendGif(chatId, file, replyToMessageId)
}

func stripName(name string) string {
	nameWithNoEmojis := string(emojiRegex.ReplaceAll([]byte(name), []byte("")))
	nameWithNoSymbols := string(symbolRegex.ReplaceAll([]byte(nameWithNoEmojis), []byte("")))
	runes := []rune(nameWithNoSymbols)
	if len(runes) > MAX_NAME_CHARACTERS {
		runes = runes[:MAX_NAME_CHARACTERS]
	}
	return strings.TrimSpace(string(runes))
}

// replyToMessageId - 0 when not a reply
func sendGif(chatId int64, file *os.File, replyToMessageId int) error {
	message := tgbotapi.NewAnimationUpload(chatId, tgbotapi.FileReader{Name: "welcome.mp4", Reader: file, Size: -1})
	message.ReplyToMessageID = replyToMessageId
	_, err := bot.Send(message)
	return err
}

func sendTextMessage(chatId int64, message string) error {
	msg := tgbotapi.NewMessage(chatId, message)
	_, err := bot.Send(msg)
	return err
}

func generateGif(username string, filename string) error {
	cmd := exec.Command(
		"ffmpeg",
		"-i",
		fmt.Sprintf(INPUT_FILE, resourceDir),
		"-vf",
		fmt.Sprintf(COMMAND, resourceDir, fontFile, username),
		"-codec:a",
		"copy",
		"-an",
		filename,
		"-y",
	)

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf(string(output))
	}

	return nil
}
