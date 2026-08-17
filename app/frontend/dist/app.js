// TarkovPilot v2 — UI. State arrives via the "state" event from Go
// and via the binding methods window.go.main.App.*

const App = () => window.go.main.App;

// errors go right onto the page: the production build has no devtools, they'd be invisible otherwise
const showError = (msg) => {
    let box = document.getElementById('err-box');
    if (!box) {
        box = document.createElement('div');
        box.id = 'err-box';
        box.style.cssText = 'background:#3a1f1f;border:1px solid #c73e3e;color:#e0b4b4;padding:8px 12px;font:12px Consolas,monospace;white-space:pre-wrap;user-select:text;';
        document.body.prepend(box);
    }
    box.textContent += msg + '\n';
};
window.addEventListener('error', (e) => showError('JS error: ' + e.message + ' @ ' + (e.filename || '') + ':' + (e.lineno || '')));
window.addEventListener('unhandledrejection', (e) => showError('Promise rejection: ' + (e.reason && (e.reason.message || e.reason))));

// --- localization ---

const i18n = {
    en: {
        updateAvailable: 'New version available',
        update: 'Update',
        connectionKey: 'Connection key',
        save: 'Save',
        checking: 'Checking key...',
        keyAccepted: 'Key accepted — connected',
        keyRejected: 'Key rejected — get a new one on the Pilot page',
        keyOffline: 'No connection to server — check internet / firewall',
        keyHint: 'Get the key on the Pilot page:',
        connNoKey: 'Not connected — enter your connection key',
        connChecking: 'Connecting...',
        connOk: 'Connected to',
        connBadKey: 'Key rejected — get a new one on the Pilot page',
        connOffline: 'No connection to server — check internet / firewall',
        statuses: 'Statuses',
        gameFolder: 'Game folder',
        gameLogs: 'Game logs',
        screenshots: 'Screenshots',
        browse: 'Browse',
        default: 'Default',
        options: 'Options',
        autostart: 'Start with Windows',
        autoclean: 'Clean session screenshots',
        autocleanHint: 'Delete screenshots taken this session when the map changes',
        region: 'Region',
        events: 'Events',
        quitHint: 'The X button minimizes to tray, full quit —',
        quit: 'Quit',
        stNoKey: 'enter connection key',
        stWatching: 'watching',
        stNotFound: 'not found',
        stGameBad: 'no Logs inside — is this really the EFT folder?',
        stLogsOff: 'not reading — check game folder',
    },
    ru: {
        updateAvailable: 'Доступна новая версия',
        update: 'Обновить',
        connectionKey: 'Ключ подключения',
        save: 'Сохранить',
        checking: 'Проверяю...',
        keyAccepted: 'Ключ принят — подключено',
        keyRejected: 'Ключ не принят — возьмите новый на странице Pilot',
        keyOffline: 'Нет связи с сервером — проверьте интернет / брандмауэр',
        keyHint: 'Ключ — на странице Pilot:',
        connNoKey: 'Не подключено — введите ключ подключения',
        connChecking: 'Подключение...',
        connOk: 'Подключено к',
        connBadKey: 'Ключ не принят — возьмите новый на странице Pilot',
        connOffline: 'Нет связи с сервером — проверьте интернет / брандмауэр',
        statuses: 'Статусы',
        gameFolder: 'Папка игры',
        gameLogs: 'Логи игры',
        screenshots: 'Скриншоты',
        browse: 'Выбрать',
        default: 'По умолчанию',
        options: 'Опции',
        autostart: 'Автозапуск с Windows',
        autoclean: 'Удалять скриншоты сессии',
        autocleanHint: 'Скриншоты, сделанные за сессию, удаляются при смене карты',
        region: 'Регион',
        events: 'События',
        quitHint: 'Крестик сворачивает в трей, полный выход —',
        quit: 'Выход',
        stNoKey: 'введите ключ подключения',
        stWatching: 'наблюдается',
        stNotFound: 'не найдена',
        stGameBad: 'внутри нет Logs — это точно папка игры?',
        stLogsOff: 'не читаются — проверьте папку игры',
    },
};

let lang = 'en';
const t = (key) => (i18n[lang] && i18n[lang][key]) || i18n.en[key] || key;

const applyLang = () => {
    document.querySelectorAll('[data-i18n]').forEach((el) => {
        el.textContent = t(el.dataset.i18n);
    });
    // hover tooltips
    document.querySelectorAll('[data-i18n-title]').forEach((el) => {
        el.title = t(el.dataset.i18nTitle);
    });
    document.getElementById('lang-en').classList.toggle('active', lang === 'en');
    document.getElementById('lang-ru').classList.toggle('active', lang === 'ru');
};

// --- state rendering ---

const $ = (id) => document.getElementById(id);

let hookIdDirty = false; // the user is editing the field — don't overwrite it with backend data
let lastState = null;

const setDot = (id, ok) => $(id).classList.toggle('ok', !!ok);

const render = (st) => {
    lastState = st;

    if (st.lang && st.lang !== lang) {
        lang = st.lang;
        applyLang();
    }

    $('version').textContent = 'v' + st.version;

    // update
    $('update-banner').classList.toggle('hidden', !st.updateReady);
    $('latest-version').textContent = st.latestVersion;

    // key
    if (!hookIdDirty && document.activeElement !== $('hookid')) {
        $('hookid').value = st.hookId || '';
    }
    $('link-site').textContent = st.host + '/pilot';

    // connection status (the bar under the header): nokey | checking | ok | badkey | offline
    const conn = st.connState || 'nokey';
    $('conn-bar').className = 'conn-bar ' + conn;
    const connTexts = {
        nokey: t('connNoKey'),
        checking: t('connChecking'),
        ok: t('connOk') + ' ' + st.host,
        badkey: t('connBadKey'),
        offline: t('connOffline'),
    };
    $('conn-text').textContent = connTexts[conn] || conn;

    // the field shows only the path, the error is a separate line below it
    setDot('dot-game', st.gameFound);
    $('note-game').textContent = st.gameFolder;
    $('err-game').textContent = t('stGameBad');
    $('err-game').classList.toggle('hidden', !!st.gameFound);

    setDot('dot-logs', st.logsWatching);
    $('note-logs').textContent = st.logsWatching ? t('stWatching') : t('stLogsOff');

    setDot('dot-screens', st.screenshotsWatching);
    $('note-screens').textContent = st.screenshotsFolder;
    $('err-screens').textContent = t('stNotFound');
    $('err-screens').classList.toggle('hidden', !!st.screenshotsWatching);

    // options
    $('chk-autostart').checked = st.autoStart;
    $('chk-autoclean').checked = st.autoClean;

    // event log
    $('event-log').textContent = (st.eventLog || []).join('\n');
};

// --- event handlers ---

const bind = () => {
    $('btn-save-key').addEventListener('click', async () => {
        const val = $('hookid').value.trim();
        const note = $('key-saved');
        const showNote = (text, isErr) => {
            note.textContent = text;
            note.style.color = isErr ? 'var(--err)' : '';
            note.classList.remove('hidden');
            setTimeout(() => { note.classList.add('hidden'); note.style.color = ''; }, 4000);
        };

        // don't save an empty key — an accidental Save would wipe a working key
        if (!val) {
            showNote(t('stNoKey'), true);
            return;
        }

        // SetHookId verifies the key on the server synchronously — wait for the result
        hookIdDirty = false;
        const btn = $('btn-save-key');
        btn.disabled = true;
        btn.textContent = t('checking');
        let st;
        try {
            st = await App().SetHookId(val);
        } finally {
            btn.disabled = false;
            btn.textContent = t('save');
        }
        render(st);

        if (st.connState === 'ok') showNote('✓ ' + t('keyAccepted'), false);
        else if (st.connState === 'badkey') showNote('✗ ' + t('keyRejected'), true);
        else showNote(t('keyOffline'), true);
    });

    $('hookid').addEventListener('input', () => { hookIdDirty = true; });

    $('link-site').addEventListener('click', (e) => {
        e.preventDefault();
        App().OpenPilotPage();
    });

    $('btn-update').addEventListener('click', async () => render(await App().DoUpdate()));

    $('btn-quit').addEventListener('click', () => App().Quit());

    // custom title bar buttons (frameless window): the X, like the native one, hides to tray
    $('btn-min').addEventListener('click', () => window.runtime.WindowMinimise());
    $('btn-close').addEventListener('click', () => window.runtime.WindowHide());

    $('btn-browse-game').addEventListener('click', async () => render(await App().BrowseGameFolder()));
    $('btn-clear-game').addEventListener('click', async () => render(await App().ClearGameFolder()));
    $('btn-browse-screens').addEventListener('click', async () => render(await App().BrowseScreenshotsFolder()));
    $('btn-clear-screens').addEventListener('click', async () => render(await App().ClearScreenshotsFolder()));

    $('chk-autostart').addEventListener('change', async (e) => render(await App().SetAutoStart(e.target.checked)));
    $('chk-autoclean').addEventListener('change', async (e) => render(await App().SetAutoClean(e.target.checked)));

    // "Region" = UI language + website (en → .com, ru → .ru).
    // Update lastState too: otherwise render(lastState) would instantly roll the choice back
    const switchRegion = async (l) => {
        lang = l;
        if (lastState) {
            lastState.lang = l;
            lastState.host = l === 'ru' ? 'tarkov-market.ru' : 'tarkov-market.com';
            render(lastState);
        }
        applyLang();
        render(await App().SetRegion(l));
    };
    $('lang-en').addEventListener('click', () => switchRegion('en'));
    $('lang-ru').addEventListener('click', () => switchRegion('ru'));
};

// --- startup ---

window.addEventListener('DOMContentLoaded', async () => {
    try {
        applyLang();
        bind();

        if (!window.go) { showError('window.go is missing'); }
        if (!window.runtime) { showError('window.runtime is missing'); }

        window.runtime.EventsOn('state', render);
        render(await App().GetState());
    } catch (e) {
        showError('init failed: ' + (e && (e.stack || e.message || e)));
    }
});
