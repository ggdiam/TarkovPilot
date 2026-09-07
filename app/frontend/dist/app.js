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
        test: 'Test',
        clear: 'Delete',
        testing: 'Testing...',
        clearing: 'Deleting...',
        checking: 'Checking key...',
        keyAccepted: 'Key accepted — connected',
        testSent: 'Data sent — activity updated on the Pilot page',
        keyRemoved: 'Connection key removed',
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
        startMinimized: 'Start minimized',
        startMinimizedHint: 'Launch directly to the tray without showing the window',
        autoclean: 'Clean session screenshots',
        autocleanHint: 'Delete screenshots taken this session when the map changes',
        region: 'Region',
        events: 'Events',
        quitHint: 'The X button minimizes to tray, full quit —',
        quit: 'Quit',
        stWatching: 'watching',
        stNotFound: 'not found',
        stGameBad: 'no Logs inside — is this really the EFT folder?',
        stLogsOff: 'not reading — check game folder',
        questSyncTitle: 'Quest progress',
        questSyncHint: 'Import completed quests from EFT logs',
        questSync: 'Sync quests',
        questSyncHeading: 'Sync completed quests',
        questSyncSince: 'EFT logs from 2026 onward',
        questSyncPrivacy: 'Pilot reads logs on this PC. Only completed quest IDs and an anonymous character key are sent.',
        openPilotPage: 'Open Pilot page',
        scanAgain: 'Scan again',
        questScanning: 'Scanning EFT logs…',
        questProfilesFound: 'Found {profiles} characters and {quests} completed quests.',
        questNoProfiles: 'No characters with completed quests were found in logs from 2026 onward.',
        questProfileTitle: '{mode} PMC #{index}',
        questCompleted: '{count} completed quests',
        questPeriod: 'Log events: {first} — {last}',
        questMatched: 'Recognized by website: {matched}; skipped: {ignored}',
        questImportExisting: 'Import into an existing PMC',
        questNoMatchingCharacters: 'No website PMC for this mode yet.',
        questCharacter: '{name} · {fraction} · level {level}',
        questLinked: 'linked',
        questImport: 'Import',
        questCreateNew: 'Or create a new PMC',
        questCreateImport: 'Create and import',
        questWorking: 'Saving quest progress…',
        questSuccess: 'Done: {imported} new, {already} already completed, {ignored} skipped.',
        questErrorConnection: 'Connect Pilot to your website account first.',
        questErrorPro: 'Quest progress is available with a PRO subscription.',
        questErrorLogs: 'EFT logs were not found. Check the game folder in Pilot.',
        questErrorNetwork: 'Could not reach the website. Check your connection and try again.',
        questErrorProfile: 'This local character is no longer available. Scan logs again.',
        questErrorCharacter: 'The selected website PMC was not found.',
        questErrorMode: 'The local character and website PMC use different game modes.',
        questErrorTarget: 'Choose a website PMC or create a new one.',
        questErrorServer: 'The website could not complete the import. Try again later.',
        modePVP: 'PVP',
        modePVE: 'PVE',
        modeSEASON: 'Seasonal PVP',
    },
    ru: {
        updateAvailable: 'Доступна новая версия',
        update: 'Обновить',
        connectionKey: 'Ключ подключения',
        save: 'Сохранить',
        test: 'Тест',
        clear: 'Удалить',
        testing: 'Проверяю...',
        clearing: 'Удаляю...',
        checking: 'Проверяю...',
        keyAccepted: 'Ключ принят — подключено',
        testSent: 'Данные отправлены — активность обновлена на странице Pilot',
        keyRemoved: 'Ключ подключения удалён',
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
        startMinimized: 'Запускать свёрнутым',
        startMinimizedHint: 'Запуск сразу в трей, без показа окна',
        autoclean: 'Удалять скриншоты сессии',
        autocleanHint: 'Скриншоты, сделанные за сессию, удаляются при смене карты',
        region: 'Регион',
        events: 'События',
        quitHint: 'Крестик сворачивает в трей, полный выход —',
        quit: 'Выход',
        stWatching: 'наблюдается',
        stNotFound: 'не найдена',
        stGameBad: 'внутри нет Logs — это точно папка игры?',
        stLogsOff: 'не читаются — проверьте папку игры',
        questSyncTitle: 'Прогресс квестов',
        questSyncHint: 'Импорт завершённых квестов из логов EFT',
        questSync: 'Синхронизировать',
        questSyncHeading: 'Синхронизация квестов',
        questSyncSince: 'Логи EFT начиная с 2026 года',
        questSyncPrivacy: 'Pilot читает логи на этом ПК. На сайт отправляются только ID завершённых квестов и анонимный ключ персонажа.',
        openPilotPage: 'Открыть страницу Pilot',
        scanAgain: 'Сканировать заново',
        questScanning: 'Сканирую логи EFT…',
        questProfilesFound: 'Найдено персонажей: {profiles}, завершённых квестов: {quests}.',
        questNoProfiles: 'В логах с 2026 года не найдено персонажей с завершёнными квестами.',
        questProfileTitle: '{mode} PMC №{index}',
        questCompleted: 'Завершено квестов: {count}',
        questPeriod: 'События в логах: {first} — {last}',
        questMatched: 'Распознано сайтом: {matched}; пропущено: {ignored}',
        questImportExisting: 'Загрузить в существующего PMC',
        questNoMatchingCharacters: 'На сайте ещё нет PMC для этого режима.',
        questCharacter: '{name} · {fraction} · уровень {level}',
        questLinked: 'привязан',
        questImport: 'Загрузить',
        questCreateNew: 'Или создать нового PMC',
        questCreateImport: 'Создать и загрузить',
        questWorking: 'Сохраняю прогресс квестов…',
        questSuccess: 'Готово: новых — {imported}, уже завершено — {already}, пропущено — {ignored}.',
        questErrorConnection: 'Сначала подключите Pilot к аккаунту на сайте.',
        questErrorPro: 'Прогресс квестов доступен по подписке PRO.',
        questErrorLogs: 'Логи EFT не найдены. Проверьте папку игры в Pilot.',
        questErrorNetwork: 'Не удалось связаться с сайтом. Проверьте соединение и повторите.',
        questErrorProfile: 'Этот локальный персонаж больше не найден. Просканируйте логи заново.',
        questErrorCharacter: 'Выбранный PMC на сайте не найден.',
        questErrorMode: 'Локальный персонаж и PMC на сайте относятся к разным режимам.',
        questErrorTarget: 'Выберите PMC на сайте или создайте нового.',
        questErrorServer: 'Сайт не смог завершить импорт. Попробуйте позже.',
        modePVP: 'PVP',
        modePVE: 'PVE',
        modeSEASON: 'Сезонный PVP',
    },
};

let lang = 'en';
const t = (key) => (i18n[lang] && i18n[lang][key]) || i18n.en[key] || key;
const tf = (key, values = {}) => Object.entries(values).reduce(
    (text, [name, value]) => text.replaceAll('{' + name + '}', String(value)),
    t(key),
);

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
    updateHookAction();
    if (questSyncState) renderQuestProfiles();
};

// --- state rendering ---

const $ = (id) => document.getElementById(id);

let hookIdDirty = false; // the user is editing the field — don't overwrite it with backend data
let lastState = null;
let hookActionBusy = false;

const setDot = (id, ok) => $(id).classList.toggle('ok', !!ok);

const savedHookId = () => ((lastState && lastState.hookId) || '').trim();
const hookAction = () => {
    const saved = savedHookId();
    const current = $('hookid').value.trim();
    if (!current) return saved ? 'clear' : 'none';
    if (current === saved) return 'test';
    return 'save';
};
const updateHookAction = () => {
    const btn = $('btn-save-key');
    if (!btn || hookActionBusy) return;
    const action = hookAction();
    btn.textContent = t(action === 'test' ? 'test' : action === 'clear' ? 'clear' : 'save');
    btn.disabled = action === 'none';
};

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
    updateHookAction();

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
    $('chk-start-minimized').checked = st.startMinimized;
    $('chk-autoclean').checked = st.autoClean;

    // event log
    $('event-log').textContent = (st.eventLog || []).join('\n');
};

// --- quest progress sync ---

let questSyncState = null;
let questSyncBusy = false;

const makeEl = (tag, className, text) => {
    const el = document.createElement(tag);
    if (className) el.className = className;
    if (text !== undefined) el.textContent = text;
    return el;
};

const questModeName = (mode) => t('mode' + mode);

const questDate = (value) => {
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) return '—';
    return date.toLocaleDateString(lang === 'ru' ? 'ru-RU' : 'en-US', {
        day: '2-digit', month: 'short', year: 'numeric',
    });
};

const setQuestStatus = (text, kind = '') => {
    const el = $('quest-sync-status');
    el.className = 'quest-sync-status' + (kind ? ' ' + kind : '');
    el.textContent = text;
};

const setQuestBusy = (busy) => {
    questSyncBusy = busy;
    document.querySelectorAll('#quest-sync-modal .quest-action').forEach((button) => {
        button.disabled = busy;
    });
};

const questErrorKey = (code) => ({
    connection_required: 'questErrorConnection',
    bad_key: 'questErrorConnection',
    pro_required: 'questErrorPro',
    logs_not_found: 'questErrorLogs',
    network_error: 'questErrorNetwork',
    profile_not_found: 'questErrorProfile',
    character_not_found: 'questErrorCharacter',
    mode_mismatch: 'questErrorMode',
    invalid_target: 'questErrorTarget',
    invalid_fraction: 'questErrorTarget',
    server_error: 'questErrorServer',
}[code] || 'questErrorServer');

const showQuestError = (code) => {
    setQuestStatus(t(questErrorKey(code)), 'error');
    const needsSite = ['connection_required', 'bad_key', 'pro_required'].includes(code);
    $('btn-quest-site').classList.toggle('hidden', !needsSite);
    $('btn-quest-scan-again').classList.toggle('hidden', code === 'pro_required');
};

const questSuccessText = (result) => tf('questSuccess', {
    imported: result ? result.importedCount : 0,
    already: result ? result.alreadyDone : 0,
    ignored: result ? result.ignoredCount : 0,
});

const runQuestImport = async (profileKey, charUid) => {
    if (questSyncBusy) return;
    setQuestBusy(true);
    setQuestStatus(t('questWorking'));
    try {
        const response = await App().ImportQuestProfile(profileKey, charUid);
        if (!response || !response.ok) {
            showQuestError(response && response.error);
            return;
        }
        questSyncState = response;
        renderQuestProfiles();
        setQuestStatus(questSuccessText(response.importResult), 'success');
    } catch (_) {
        showQuestError('network_error');
    } finally {
        setQuestBusy(false);
    }
};

const runQuestCreate = async (profileKey, fraction) => {
    if (questSyncBusy) return;
    setQuestBusy(true);
    setQuestStatus(t('questWorking'));
    try {
        const response = await App().CreateQuestCharacterAndImport(profileKey, fraction);
        if (!response || !response.ok) {
            showQuestError(response && response.error);
            return;
        }
        questSyncState = response;
        renderQuestProfiles();
        setQuestStatus(questSuccessText(response.importResult), 'success');
    } catch (_) {
        showQuestError('network_error');
    } finally {
        setQuestBusy(false);
    }
};

const renderQuestProfiles = () => {
    const root = $('quest-sync-results');
    root.replaceChildren();

    const profiles = (questSyncState && questSyncState.profiles) || [];
    if (!profiles.length) {
        root.classList.add('hidden');
        setQuestStatus(t('questNoProfiles'), 'warning');
        return;
    }

    const modeIndexes = {};
    profiles.forEach((profile) => {
        modeIndexes[profile.gameMode] = (modeIndexes[profile.gameMode] || 0) + 1;
        const card = makeEl('article', 'quest-profile');
        const head = makeEl('div', 'quest-profile-head');
        const title = makeEl('div', 'quest-profile-title', tf('questProfileTitle', {
            mode: questModeName(profile.gameMode),
            index: modeIndexes[profile.gameMode],
        }));
        const count = makeEl('div', 'quest-count', tf('questCompleted', { count: profile.foundCount }));
        head.append(title, count);
        card.append(head);
        card.append(makeEl('div', 'quest-meta', tf('questPeriod', {
            first: questDate(profile.firstEventAt),
            last: questDate(profile.lastEventAt),
        })));
        card.append(makeEl('div', 'quest-meta', tf('questMatched', {
            matched: profile.matchedCount,
            ignored: profile.ignoredCount,
        })));

        const existing = makeEl('div', 'quest-target');
        existing.append(makeEl('div', 'quest-target-label', t('questImportExisting')));
        const existingRow = makeEl('div', 'quest-target-row');
        const select = makeEl('select', 'quest-character-select');
        const characters = ((questSyncState && questSyncState.characters) || [])
            .filter((character) => character.gameMode === profile.gameMode);
        characters.forEach((character) => {
            const label = tf('questCharacter', {
                name: character.charName || character.fraction || 'PMC',
                fraction: character.fraction || 'PMC',
                level: character.level || 1,
            }) + (character.uid === profile.linkedCharUid ? ' · ' + t('questLinked') : '');
            const option = makeEl('option', '', label);
            option.value = character.uid;
            option.selected = character.uid === profile.linkedCharUid;
            select.append(option);
        });
        const importButton = makeEl('button', 'quest-action', t('questImport'));
        importButton.disabled = characters.length === 0 || questSyncBusy;
        importButton.addEventListener('click', () => runQuestImport(profile.profileKey, select.value));
        if (characters.length) {
            existingRow.append(select, importButton);
        } else {
            existingRow.append(makeEl('div', 'quest-no-target', t('questNoMatchingCharacters')));
        }
        existing.append(existingRow);
        card.append(existing);

        const create = makeEl('div', 'quest-target create');
        create.append(makeEl('div', 'quest-target-label', t('questCreateNew')));
        const createRow = makeEl('div', 'quest-target-row');
        const fraction = makeEl('select', 'quest-fraction-select');
        for (const [value, label] of [['Bear', 'BEAR'], ['USEC', 'USEC']]) {
            const option = makeEl('option', '', label);
            option.value = value;
            fraction.append(option);
        }
        const createButton = makeEl('button', 'quest-action', t('questCreateImport'));
        createButton.disabled = questSyncBusy;
        createButton.addEventListener('click', () => runQuestCreate(profile.profileKey, fraction.value));
        createRow.append(fraction, createButton);
        create.append(createRow);
        card.append(create);

        root.append(card);
    });
    root.classList.remove('hidden');
};

const startQuestScan = async () => {
    if (questSyncBusy) return;
    $('btn-quest-site').classList.add('hidden');
    $('btn-quest-scan-again').classList.add('hidden');
    $('quest-sync-results').classList.add('hidden');
    questSyncState = null;

    if (!lastState || lastState.connState !== 'ok') {
        showQuestError('connection_required');
        return;
    }
    if (!lastState.pro) {
        showQuestError('pro_required');
        return;
    }

    setQuestBusy(true);
    setQuestStatus(t('questScanning'), 'loading');
    try {
        const response = await App().ScanQuestHistory();
        if (!response || !response.ok) {
            showQuestError(response && response.error);
            return;
        }
        questSyncState = response;
        renderQuestProfiles();
        const profiles = response.profiles || [];
        if (profiles.length) {
            const quests = profiles.reduce((sum, profile) => sum + (profile.foundCount || 0), 0);
            setQuestStatus(tf('questProfilesFound', { profiles: profiles.length, quests }), 'success');
        }
        $('btn-quest-scan-again').classList.remove('hidden');
    } catch (_) {
        showQuestError('network_error');
    } finally {
        setQuestBusy(false);
    }
};

const openQuestSync = () => {
    $('quest-sync-modal').classList.remove('hidden');
    startQuestScan();
};

const closeQuestSync = () => {
    $('quest-sync-modal').classList.add('hidden');
};

// --- event handlers ---

const bind = () => {
    $('btn-save-key').addEventListener('click', async () => {
        const val = $('hookid').value.trim();
        const action = hookAction();
        const isTest = action === 'test';
        const isClear = action === 'clear';
        const note = $('key-saved');
        const showNote = (text, isErr) => {
            note.textContent = text;
            note.style.color = isErr ? 'var(--err)' : '';
            note.classList.remove('hidden');
            setTimeout(() => { note.classList.add('hidden'); note.style.color = ''; }, 4000);
        };

        if (action === 'none') return;

        // Saving verifies a changed key; testing sends status; clearing removes the saved key locally.
        hookIdDirty = false;
        const btn = $('btn-save-key');
        hookActionBusy = true;
        btn.disabled = true;
        btn.textContent = t(isTest ? 'testing' : isClear ? 'clearing' : 'checking');
        let st;
        try {
            st = isTest ? await App().TestConnection() : await App().SetHookId(val);
        } catch (_) {
            showNote(t('keyOffline'), true);
            return;
        } finally {
            hookActionBusy = false;
            btn.disabled = false;
            updateHookAction();
        }
        render(st);

        if (isClear) showNote('✓ ' + t('keyRemoved'), false);
        else if (st.connState === 'ok') showNote('✓ ' + t(isTest ? 'testSent' : 'keyAccepted'), false);
        else if (st.connState === 'badkey') showNote('✗ ' + t('keyRejected'), true);
        else showNote(t('keyOffline'), true);
    });

    $('hookid').addEventListener('input', () => {
        hookIdDirty = $('hookid').value.trim() !== savedHookId();
        updateHookAction();
    });

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
    $('chk-start-minimized').addEventListener('change', async (e) => render(await App().SetStartMinimized(e.target.checked)));
    $('chk-autoclean').addEventListener('change', async (e) => render(await App().SetAutoClean(e.target.checked)));

    $('btn-quest-sync').addEventListener('click', openQuestSync);
    $('btn-quest-sync-close').addEventListener('click', closeQuestSync);
    $('btn-quest-scan-again').addEventListener('click', startQuestScan);
    $('btn-quest-site').addEventListener('click', () => App().OpenPilotPage());
    $('quest-sync-modal').addEventListener('click', (e) => {
        if (e.target === $('quest-sync-modal')) closeQuestSync();
    });
    window.addEventListener('keydown', (e) => {
        if (e.key === 'Escape') closeQuestSync();
    });

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
