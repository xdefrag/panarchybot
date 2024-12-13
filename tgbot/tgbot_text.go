package tgbot

const (
	textError                  = "Все поломалось. Пинаю разработчиков, попробуйте чуть позже."
	textFollowUpButton         = "Забирай эйрдроп"
	textFollowUpLink           = "https://t.me/panarchybot?start"
	textThanksErrorAmount      = "Сорян, не распарсил количество :("
	textThanksErrorAccountFrom = "Не найден аккаунт отправителя. Тыкай давай в @panarchybot"
	textThanksErrorAccountTo   = "Не найден аккаунт получателя. Пусть тыкнет в @panarchybot"
	textTemplateThanksSuccess  = "Успешно отправлено %f пользователю @%s."
	textSuggestWelcome         = "Предложи новость или задай вопрос."
	textSuggestSubmited        = "Спасибо за предложку!"
	textSendErrorUserNotFound  = "Пользователь не найден."
	textSendToWhom             = "Кому отправить?"
	textSendAmount             = "Сколько отправить?"
	textSendYes                = "Отправить"
	textSendNo                 = "Отмена"
	textSendErrorNotEnough     = "Сумма больше баланса."
	textSendNotifySendingTx    = "Отправляю транзакцию..."
	textSendError              = "Ошибка отправки транзакции."
	textTemplateSendSuccess    = "<a href=\"%s\">Транзакция</a> успешно отправлена"
	textTemplateSendConfirm    = "Отправить <a href=\"%s%s\">%s</a> %s %s?"
	textStartSend              = "Отправить"
	textStartSuggest           = "Предложить"
	textStartRegister          = "Регистрация"
	textStartWelcome           = `Приветствую, свободный человек!

Этот бот предоставляет доступ для взаимодействия с каналом <a href="%[1]s">%[2]s</a>. Перед взаимодействием необходимо пройти регистрацию. Во время регистрации будет создан счет на блокчейне Stellar и на него отправится эйрдроп в виде токенов %[3]s. Их можно отправить другим подписчикам канала как благодарность за развернутый и полезный ответ или просто так.

Также можно предложить публикацию или задать вопрос автору канала.

Нажми «Регистриация», чтобы создать свой счет.
`
	textStartDashboard = `Счет <a href="%s%s">%s</a>
%s %s

Нажми «Отправить», чтобы отправить токены %[5]s.
Нажми «Предложить», чтобы предложить публикацию или задать вопрос.`
)
