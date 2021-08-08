package Senko

import (
	"bytes"
	"encoding/json"
	"fmt"
	tgbotapi "github.com/4Kaze/telegram-bot-api/v5"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	m.Run()
}

func TestHandleRequestWithNewMember(t *testing.T) {
	//given
	setupTest(t)
	messageId := 10

	//when
	response := newMemberJoins(messageId)

	//then
	assertResponseStatusWasOk(t, response)
	assertOneMessageWasSent(t).toChat(TEST_CHAT_ID).asReplyTo(messageId).withMp4Animation().withFileSizeAbove(200 * kB)
}

func TestHandleRegularGroupMessage(t *testing.T) {
	//given
	setupTest(t)

	//when
	response := groupChatMessageIsReceived("Some message")

	//then
	assertResponseStatusWasOk(t, response)
	assertNoMessageWasSent(t)
}

func TestHandleGroupCommands(t *testing.T) {
	commands := []string{"/start", "/wepo", "/genewate"}
	for _, command := range commands {
		//given
		setupTest(t)

		//when
		response := groupChatMessageIsReceived(command)

		//then
		assertResponseStatusWasOk(t, response)
		assertNoMessageWasSent(t)
	}
}

func TestHandleRegularPrivateChatMessage(t *testing.T) {
	//given
	setupTest(t)

	//when
	response := privateChatMessageIsReceived("Some message")

	//then
	assertResponseStatusWasOk(t, response)
	assertNoMessageWasSent(t)
}

func TestHandlePrivateChatCommands(t *testing.T) {
	cases := []struct {
		command       string
		expectedReply string
	}{
		{
			command:       "/start",
			expectedReply: START_REPLY,
		},
		{
			command:       "/start   something",
			expectedReply: START_REPLY,
		},
		{
			command:       "/wepo",
			expectedReply: REPO_URL,
		},
		{
			command:       "/genewate",
			expectedReply: GENERATE_USAGE,
		},
		{
			command:       "/genewate    ",
			expectedReply: GENERATE_USAGE,
		},
	}

	for _, tc := range cases {
		//given
		setupTest(t)

		//when
		response := privateChatMessageIsReceived(tc.command)

		//then
		assertResponseStatusWasOk(t, response)
		assertOneMessageWasSent(t).toChat(TEST_CHAT_ID).withText(tc.expectedReply)
	}
}

func TestHandlePrivateChatGenerateGif(t *testing.T) {
	//given
	setupTest(t)

	//when
	response := privateChatMessageIsReceived("/genewate some name")

	//then
	assertResponseStatusWasOk(t, response)
	assertNMessagesWereSent(t, 2)
	assertTextMessageWasSent(t, GENERATION_STARTED_MESSAGE)
	assertMp4AnimationWasSent(t)
}

func TestStripName(t *testing.T) {
	cases := []struct {
		inputName      string
		expectedOutput string
	}{
		{
			inputName:      "Name",
			expectedOutput: "Name",
		},
		{
			inputName:      "Some name",
			expectedOutput: "Some name",
		},
		{
			inputName:      strings.Repeat("A", 20),
			expectedOutput: strings.Repeat("A", 20),
		},
		{
			inputName:      strings.Repeat("A", 21),
			expectedOutput: strings.Repeat("A", 20),
		},
		{
			inputName:      "Em😊ji name",
			expectedOutput: "Emji name",
		},
		{
			inputName:      strings.Repeat("😊", 20) + " Long name",
			expectedOutput: "Long name",
		},
		{
			inputName:      "':N?am^e$#@&*()",
			expectedOutput: "Name",
		},
	}

	for _, tc := range cases {
		//when
		actualOutput := stripName(tc.inputName)

		//then
		assert.Equal(t, tc.expectedOutput, actualOutput, "Expected different outcome when stripping a name")
	}
}

func TestHandleWrongRequest(t *testing.T) {
	//given
	setupTest(t)

	//when
	response := makeRequest("{'some': 'wrong request'}")

	//then
	assertResponseStatusWasOk(t, response)
	assertNoMessageWasSent(t)
}

func TestMessageValidation(t *testing.T) {
	messages := []tgbotapi.Message{
		{
			MessageID: 1,
			Chat:      nil,
			NewChatMembers: &[]tgbotapi.User{
				{ID: 0, FirstName: "Some name"},
			},
		},
		{
			MessageID: 1,
			Chat: &tgbotapi.Chat{
				ID:   TEST_CHAT_ID,
				Type: "channel",
			},
			NewChatMembers: &[]tgbotapi.User{
				{ID: 0, FirstName: "Some name"},
			},
		},
		{
			MessageID: 1,
			Chat: &tgbotapi.Chat{
				ID:   TEST_CHAT_ID,
				Type: "private",
			},
			NewChatMembers: &[]tgbotapi.User{
				{ID: 0, FirstName: "Some name"},
			},
		},
		{
			MessageID: 1,
			Chat: &tgbotapi.Chat{
				ID:   TEST_CHAT_ID,
				Type: "group",
			},
			NewChatMembers: &[]tgbotapi.User{},
		},
	}

	for _, message := range messages {
		//given
		setupTest(t)

		//when
		response := makeRequest(tgbotapi.Update{Message: &message})

		//then
		assertResponseStatusWasOk(t, response)
		assertNoMessageWasSent(t)
	}
}

// ========= ABILITIES ==========

func newMemberJoins(messageId int) *httptest.ResponseRecorder {
	update := tgbotapi.Update{
		UpdateID: 0,
		Message: &tgbotapi.Message{
			MessageID: messageId,
			Chat: &tgbotapi.Chat{
				ID:   TEST_CHAT_ID,
				Type: "group",
			},
			NewChatMembers: &[]tgbotapi.User{
				{ID: 0, FirstName: "Some name"},
			},
		},
	}
	return makeRequest(update)
}

func privateChatMessageIsReceived(text string) *httptest.ResponseRecorder {
	return messageReceived(text, "private")
}

func groupChatMessageIsReceived(text string) *httptest.ResponseRecorder {
	return messageReceived(text, "group")
}

func messageReceived(text string, groupType string) *httptest.ResponseRecorder {
	update := tgbotapi.Update{
		UpdateID: 0,
		Message: &tgbotapi.Message{
			Chat: &tgbotapi.Chat{
				ID:   TEST_CHAT_ID,
				Type: groupType,
			},
			Text: text,
		},
	}
	return makeRequest(update)
}

func makeRequest(body interface{}) *httptest.ResponseRecorder {
	jsonBuf, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/", bytes.NewReader(jsonBuf))
	req.Header.Add("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	HandleRequest(resp, req)
	return resp
}

// ========= ASSERTIONS ==========

func assertNoMessageWasSent(t *testing.T) {
	assert.Equal(t, 0, len(sentMessages), "Expected no messages to be sent")
}

func assertOneMessageWasSent(t *testing.T) *MessageAssertion {
	assert.Equal(t, 1, len(sentMessages), "Expected exactly one message to be sent")
	return &MessageAssertion{
		t:       t,
		message: sentMessages[0],
	}
}

func assertNMessagesWereSent(t *testing.T, expectedNumberOfMessages int) {
	assert.Equal(t, expectedNumberOfMessages, len(sentMessages))
}

func assertTextMessageWasSent(t *testing.T, text string) {
	for _, message := range sentMessages {
		if message.text == text && message.chatId == TEST_CHAT_ID {
			return
		}
	}
	t.Errorf("A message with text %v not found", text)
}

func assertMp4AnimationWasSent(t *testing.T) {
	for _, message := range sentMessages {
		if message.file.size > 0 && len(message.file.name) > 0 && message.file.mimeType == MP4_MIME_TYPE {
			return
		}
	}
	t.Errorf("No message with a mp4 animation was found")
}

func assertResponseStatusWasOk(t *testing.T, resp *httptest.ResponseRecorder) {
	assert.Equal(t, 200, resp.Code, "Expected response status to be 200")
}

type MessageAssertion struct {
	t       *testing.T
	message Message
}

func (messageAssertion *MessageAssertion) toChat(chatId int64) *MessageAssertion {
	assert.Equal(messageAssertion.t, chatId, messageAssertion.message.chatId, "Expected message to be sent with different chatId")
	return messageAssertion
}

func (messageAssertion *MessageAssertion) withText(text string) *MessageAssertion {
	assert.Equal(messageAssertion.t, text, messageAssertion.message.text, "Expected message to be sent with different text")
	return messageAssertion
}

func (messageAssertion *MessageAssertion) asReplyTo(messageId int) *MessageAssertion {
	assert.Equal(messageAssertion.t, messageId, messageAssertion.message.replyToMessage, "Expected message to be sent as a reply")
	return messageAssertion
}

func (messageAssertion *MessageAssertion) withMp4Animation() *MessageAssertion {
	assert.Equal(messageAssertion.t, MP4_MIME_TYPE, messageAssertion.message.file.mimeType, "Expected message to have mp4 animation attached")
	return messageAssertion
}

func (messageAssertion *MessageAssertion) withFileSizeAbove(filesize int) *MessageAssertion {
	assert.Greater(messageAssertion.t, messageAssertion.message.file.size, filesize, "Expected message to have mp4 file with proper size")
	return messageAssertion
}

// ========= STUBS ==========

func stubGetMe() {
	httpmock.RegisterResponder("POST", telegramUrl("getMe"), httpmock.NewStringResponder(200, GET_ME_RESPONSE))
}

func stubSendMessage() {
	httpmock.RegisterResponder("POST", telegramUrl("sendMessage"),
		func(r *http.Request) (*http.Response, error) {
			r.ParseForm()
			chatId, _ := strconv.ParseInt(r.Form.Get("chat_id"), 10, 64)
			replyMessageId, _ := strconv.Atoi(r.Form.Get("reply_to_message_id"))
			sentMessages = append(sentMessages, Message{
				chatId:         chatId,
				replyToMessage: replyMessageId,
				text:           r.Form.Get("text"),
			})
			return httpmock.NewStringResponse(200, SUCCESSFUL_RESPONSE), nil
		},
	)
}

func stubSendAnimation() {
	httpmock.RegisterResponder("POST", telegramUrl("sendAnimation"),
		func(r *http.Request) (*http.Response, error) {
			r.ParseForm()
			file, _ := extractFileInfo(r, "animation")
			chatId, _ := strconv.ParseInt(r.Form.Get("chat_id"), 10, 64)
			replyMessageId, _ := strconv.Atoi(r.Form.Get("reply_to_message_id"))
			sentMessages = append(sentMessages, Message{
				chatId:         chatId,
				replyToMessage: replyMessageId,
				file:           *file,
			})
			return httpmock.NewStringResponse(200, SUCCESSFUL_RESPONSE), nil
		},
	)
}

// ========= SETUPS ==========

func setupTest(t *testing.T) {
	httpmock.Reset()
	sentMessages = nil
	stubGetMe()
	stubSendMessage()
	stubSendAnimation()
}

func init() {
	_ = os.Setenv("TOKEN", TEST_TOKEN)
	resourceDir = "."
	fontFile = "TestFont.otf"
	sentMessages = make([]Message, 0)
}

// ========= HELPERS ==========

func telegramUrl(path string) string {
	return fmt.Sprintf(TELEGRAM_ENDPOINT, TEST_TOKEN, path)
}

func extractFileInfo(r *http.Request, fieldName string) (*GifFile, error) {
	file, fileHeader, err := r.FormFile(fieldName)
	if err != nil {
		return nil, err
	}
	fileHeaderBuffer := make([]byte, 512)
	if _, err := file.Read(fileHeaderBuffer); err != nil {
		return nil, err
	}

	return &GifFile{
		name:     fileHeader.Filename,
		size:     int(fileHeader.Size),
		mimeType: http.DetectContentType(fileHeaderBuffer),
	}, nil
}

type Message struct {
	chatId         int64
	replyToMessage int
	text           string
	file           GifFile
}

type GifFile struct {
	name     string
	size     int
	mimeType string
}

var sentMessages []Message

const kB = 1024

// ========= DATA ==========

const (
	TEST_TOKEN        = "testToken123"
	TEST_CHAT_ID      = 0
	TELEGRAM_ENDPOINT = "https://api.telegram.org/bot%s/%s"
	GET_ME_RESPONSE   = `{
		"ok": true,
		"result": {
			"id": 2137,
			"is_bot": true,
			"first_name": "Senko-San",
			"username": "senkosanbot",
			"can_join_groups": true,
			"can_read_all_group_messages": true,
			"supports_inline_queries": false
		}
	}`
	SUCCESSFUL_RESPONSE = `{
		"ok": true,
		"result": {}
	}`
	MP4_MIME_TYPE = "video/mp4"
)
