package hubview

// hubScripts contains all modular client-side JavaScript for live polling, dual-face rotation, and level up spotlight
const hubScripts = `
    function toggleCardFace(key) {
      var card = document.getElementById('card-' + key);
      if (card) {
        card.classList.toggle('is-flipped');
      }
    }

    var autoFlipCycle = 0;
    setInterval(function() {
      autoFlipCycle = 1 - autoFlipCycle;
      var cards = document.querySelectorAll('.bot-card:not(:hover)');
      cards.forEach(function(card) {
        if (autoFlipCycle === 1) {
          card.classList.add('is-flipped');
        } else {
          card.classList.remove('is-flipped');
        }
      });
    }, 8500);

    function formatNumber(n) {
      if (n < 1000) return n.toString();
      if (n < 1000000) return (n / 1000).toFixed(1) + 'k';
      return (n / 1000000).toFixed(2) + 'kk';
    }

    function formatRate(n) {
      if (!n || n <= 0) return 'Calculando...';
      return '+' + formatNumber(Math.round(n)) + '/h';
    }

    function formatUptime(sec, fallback) {
      if (sec && sec > 0) {
        var h = Math.floor(sec / 3600);
        var m = Math.floor((sec % 3600) / 60);
        return h + 'h ' + (m < 10 ? '0' : '') + m + 'm';
      }
      return fallback || '24/7 Ativo';
    }

    // Catálogo de GIFs animados temáticos locais dos personagens por jogo
    var gameCelebrationGifs = {
      'pokelegacy': 'assets/gifs/pokelegacy.gif', 'pokeidle': 'assets/gifs/pokeidle.gif',
      'pokeidleworld': 'assets/gifs/pokeidleworld.gif', 'pokepixel': 'assets/gifs/pokepixel.gif',
      'poketrue': 'assets/gifs/poketrue.gif', 'huntera': 'assets/gifs/huntera.gif',
      'tibidle': 'assets/gifs/tibidle.gif', 'baiak': 'assets/gifs/baiak.gif',
      'luxidle': 'assets/gifs/luxidle.gif', 'lolidle': 'assets/gifs/lolidle.gif',
      'leveling': 'assets/gifs/leveling.gif', 'sigma': 'assets/gifs/sigma.gif',
      'farmricochet': 'assets/gifs/farmricochet.gif', 'default': 'assets/gifs/pokelegacy.gif'
    };

    // Fullscreen Level Up Spotlight System
    var knownLevels = {}, levelUpQueue = [], isShowingLevelUp = false, levelUpTimer = null;

    function triggerLevelUpModal(info) {
      var dup = levelUpQueue.some(function(it) { return it.botKey === info.botKey && it.newLevel === info.newLevel; });
      if (dup) return;
      levelUpQueue.push(info);
      if (!isShowingLevelUp) showNextLevelUp();
    }

    function showNextLevelUp() {
      if (levelUpQueue.length === 0) { isShowingLevelUp = false; return; }
      isShowingLevelUp = true;
      var item = levelUpQueue.shift();

      var overlay = document.getElementById('levelUpOverlay');
      var gameEl = document.getElementById('lvlGameName');
      var oldLvlEl = document.getElementById('lvlOldVal');
      var newLvlEl = document.getElementById('lvlNewVal');
      var charEl = document.getElementById('lvlCharInfo');
      var areaEl = document.getElementById('lvlAreaInfo');
      var timerFill = document.getElementById('lvlTimerFill');
      var gifEl = document.getElementById('lvlCharGif');
      var fallbackEl = document.getElementById('lvlFallbackEmoji');

      if (gameEl) gameEl.textContent = (item.icon ? item.icon + ' ' : '') + item.gameName;
      if (oldLvlEl) oldLvlEl.textContent = 'Nv. ' + item.oldLevel;
      if (newLvlEl) newLvlEl.textContent = 'Nv. ' + item.newLevel;
      if (charEl) charEl.textContent = '⚔️ ' + item.charName + (item.vocation ? ' (' + item.vocation + ')' : '');
      if (areaEl) areaEl.textContent = item.area ? '🗺️ ' + item.area : '⚡ Campo de Batalha';

      var targetGif = item.gif || gameCelebrationGifs[item.botKey] || gameCelebrationGifs['default'];
      if (gifEl) { gifEl.style.display = 'block'; gifEl.src = targetGif; }
      if (fallbackEl) fallbackEl.style.display = 'none';

      if (timerFill) {
        timerFill.style.animation = 'none';
        void timerFill.offsetWidth;
        timerFill.style.animation = 'shrinkTimer 4s linear forwards';
      }
      if (overlay) overlay.classList.add('active');

      if (levelUpTimer) clearTimeout(levelUpTimer);
      levelUpTimer = setTimeout(function() { dismissLevelUp(); }, 4000);
    }

    function dismissLevelUp() {
      if (levelUpTimer) { clearTimeout(levelUpTimer); levelUpTimer = null; }
      var overlay = document.getElementById('levelUpOverlay');
      if (overlay) overlay.classList.remove('active');
      setTimeout(function() { showNextLevelUp(); }, 400);
    }

    var testGames = ['pokelegacy', 'huntera', 'tibidle', 'luxidle', 'pokepixel', 'baiak', 'farmricochet', 'leveling', 'poketrue', 'pokeidle', 'pokeidleworld', 'lolidle', 'sigma'];
    var testMeta = {
      'pokelegacy': { c: 'Charizard Flamethrower', v: 'Fire Dragon', a: 'Cinnabar Volcano' },
      'huntera': { c: 'Sir Lancelot', v: 'Paladin Knight', a: 'Dungeon of Valor' },
      'tibidle': { c: 'Archmage Gandalf', v: 'Elder Sorcerer', a: 'Carlin Magic Academy' },
      'luxidle': { c: 'Sylveon Fairy Spark', v: 'Fairy Sorceress', a: 'Demacia Woods' },
      'pokepixel': { c: 'Gengar Shadow', v: 'Ghost Shadow', a: 'Lavender Ghost Tower' },
      'baiak': { c: 'Demon Slayer', v: 'Elite Knight', a: 'Annihilator Pit' },
      'farmricochet': { c: 'Torchic Farm', v: 'Golden Rooster', a: 'Bouncing Cornfields' },
      'leveling': { c: 'Shadow Monarch Jin-Woo', v: 'Solo Hunter', a: 'Rank-S Red Gate' },
      'poketrue': { c: 'Mewtwo Psystrike', v: 'Psychic Legend', a: 'Cerulean Unknown Cave' },
      'pokeidle': { c: 'Pikachu Volt Tackle', v: 'Electric Mouse', a: 'Power Plant' },
      'pokeidleworld': { c: 'Blastoise Hydro Pump', v: 'Water Fortress', a: 'Seafoam Islands' },
      'lolidle': { c: 'Lucario Aura Sphere', v: 'Aura Champion', a: 'Summoner Rift' },
      'sigma': { c: 'Deoxys Cyber Speed', v: 'Cosmic Alien', a: 'Space Observatory' }
    };
    var testIdx = 0;
    function testLevelUp() {
      var k = testGames[testIdx % testGames.length];
      var m = testMeta[k] || { c: 'Herói', v: 'Aventureiro', a: 'Mundo Aberto' };
      testIdx++;
      triggerLevelUpModal({
        botKey: k,
        icon: k.startsWith('poke') ? '🐉' : (k === 'luxidle' ? '✨' : (k === 'farmricochet' ? '🌾' : '⚔️')),
        gameName: k.toUpperCase(),
        charName: m.c + ' (Nv. ' + (80 + testIdx) + ')',
        vocation: m.v,
        area: m.a,
        oldLevel: 79 + testIdx,
        newLevel: 80 + testIdx,
        gif: gameCelebrationGifs[k]
      });
    }
    window.testLevelUp = testLevelUp;

    window.addEventListener('keydown', function(e) {
      if ((e.key === 'l' || e.key === 'L') && !e.target.matches('input, textarea')) {
        testLevelUp();
      }
    });

    function getUnifiedAccounts(b) {
      if (!b || !b.status) return [];

      // Kick Miner
      if (b.status.accounts && typeof b.status.accounts === 'object') {
        var accKeys = Object.keys(b.status.accounts);
        if (accKeys.length > 0 && b.status.accounts[accKeys[0]] && b.status.accounts[accKeys[0]].streamers) {
          return accKeys.sort().map(function(k) {
            var entry = b.status.accounts[k];
            var streams = entry.streamers || {};
            var sKeys = Object.keys(streams).sort();
            var totalPts = 0;
            var liveCount = 0;
            var liveNames = [];
            for (var si = 0; si < sKeys.length; si++) {
              var st = streams[sKeys[si]];
              totalPts += (st.points || 0);
              if (st.isLive) {
                liveCount++;
                if (liveNames.length < 3) liveNames.push(st.streamer || sKeys[si]);
              }
            }
            var areaText = (entry.activeStreams || liveCount) + ' canais ativos';
            if (liveNames.length > 0) areaText = 'Ao vivo: ' + liveNames.join(', ') + ' (+' + liveCount + ')';
            return {
              name: 'Conta Kick #1',
              level: 0,
              vocation: 'Kick Stream',
              currentArea: areaText,
              gold: totalPts,
              kills: liveCount,
              status: 'Minerando',
              isStream: true,
              durationSec: 0
            };
          });
        }
      }

      if (Array.isArray(b.status.accounts) && b.status.accounts.length > 0) {
        return b.status.accounts;
      }
      if (b.status.accounts && typeof b.status.accounts === 'object' && Object.keys(b.status.accounts).length > 0) {
        return Object.values(b.status.accounts);
      }
      if (b.status.bots && typeof b.status.bots === 'object') {
        var keys = Object.keys(b.status.bots).sort();
        return keys.map(function(k) {
          var entry = b.status.bots[k];
          var s = entry.status || {};
          var tel = s.telemetry || {};
          var rates = s.rates || {};
          var statusStr = s.status || 'online';
          if (s.treino) statusStr = 'Treino';
          return {
            name: entry.label || entry.name || k,
            level: s.level || tel.level || 1,
            vocation: s.vocation || '',
            currentArea: s.currentArea || (s.hunt && s.hunt.name) || (s.treino ? 'Treino de Skills' : ''),
            gold: s.gold || tel.gold || 0,
            kills: s.kills || tel.kills || 0,
            stamina: tel.stamina || '',
            status: statusStr,
            durationSec: s.uptimeSec || 0,
            expPerHour: rates.xpPerHour || 0,
            goldPerHour: rates.goldPerHour || 0
          };
        });
      }
      if (b.status.statuses && typeof b.status.statuses === 'object') {
        var keys = Object.keys(b.status.statuses).sort();
        return keys.map(function(k) {
          var entry = b.status.statuses[k];
          var streams = entry.streamers || {};
          var watching = Object.values(streams).filter(function(x) { return x.watching; }).map(function(x) { return x.displayName || x.login; });
          var liveCount = Object.values(streams).filter(function(x) { return x.isLive; }).length;
          var totalPts = 0;
          Object.values(streams).forEach(function(x) { totalPts += (x.points || 0); });
          return {
            name: entry.username || k,
            level: 0,
            vocation: 'Twitch Drops',
            currentArea: watching.length ? 'Assistindo ' + watching.join(', ') : 'Minerando canais Twitch',
            gold: totalPts,
            kills: liveCount,
            status: 'Online',
            isStream: true,
            durationSec: 0
          };
        });
      }

      if (b.status.uptime || b.status.ramSysMB || b.status.goroutines) {
        return [{
          name: 'Sessão Conectada',
          level: 1,
          vocation: 'Operação Contínua',
          currentArea: 'Monitorando em Standby',
          status: 'Standby',
          durationSec: 0
        }];
      }

      return [];
    }

    async function refreshHub() {
      try {
        var res = await fetch('status.json?t=' + Date.now());
        if (!res.ok) return;
        var data = await res.json();
        if (!data || !data.bots) return;

        if (data.vps) {
          var v = data.vps;
          var ramVal = document.getElementById('ramVal');
          var ramProg = document.getElementById('ramProgress');
          var ramSub = document.getElementById('ramSub');
          if (ramVal) ramVal.textContent = (v.ramUsedMb || 0) + ' / ' + (v.ramTotalMb || 0) + ' MB';
          if (ramProg) ramProg.style.width = (v.ramPercent || 0) + '%';
          if (ramSub) ramSub.textContent = 'Livre: ' + (v.ramAvailableMb || 0) + ' MB (' + (v.ramPercent || 0) + '% em uso)';
        }

        var botKeys = Object.keys(data.bots);
        var activeCount = botKeys.filter(function(k) { return getUnifiedAccounts(data.bots[k]).length > 0; }).length;
        var botsCountEl = document.getElementById('botsCountVal');
        if (botsCountEl) botsCountEl.textContent = activeCount + ' / ' + botKeys.length;

        for (var i = 0; i < botKeys.length; i++) {
          var k = botKeys[i];
          var b = data.bots[k];
          var accList = document.getElementById('acclist-' + k);
          var badge = document.getElementById('badge-' + k);
          var accounts = getUnifiedAccounts(b);
          var isStream = (k === 'kickminer' || k === 'twitchminer');
          var isService = (k === 'pocketweb' || k === 'kaggle' || k === 'telegram');

          // Verificação de Level Up com chave única por conta e sem repetição
          for (var ai = 0; ai < accounts.length; ai++) {
            var a = accounts[ai];
            var curLvl = parseInt(a.level) || 0;
            var accKey = k + '_' + ai + (a.name ? '_' + a.name : '');
            if (knownLevels[accKey] === undefined) {
              knownLevels[accKey] = curLvl;
            } else if (curLvl > knownLevels[accKey]) {
              if (knownLevels[accKey] > 0 && !isStream && !isService) {
                triggerLevelUpModal({
                  botKey: k,
                  icon: b.icon || '⚔️',
                  gameName: b.name || b.game || k,
                  charName: a.name || ('Conta #' + (ai + 1)),
                  vocation: a.vocation || '',
                  area: a.currentArea || '',
                  oldLevel: knownLevels[accKey],
                  newLevel: curLvl,
                  gif: gameCelebrationGifs[k] || gameCelebrationGifs['default']
                });
              }
              knownLevels[accKey] = curLvl;
            }
          }

          if (badge && accounts.length > 0) {
            if (isStream) {
              var wave = '<span class="live-wave"><span></span><span></span><span></span></span>';
              badge.innerHTML = wave + (accounts[0].kills > 0 ? accounts[0].kills + ' AO VIVO' : 'MINERANDO');
            } else if (isService) {
              badge.textContent = accounts[0].status ? accounts[0].status.toUpperCase() : 'ONLINE';
            } else {
              badge.textContent = accounts.length > 1 ? accounts.length + '/' + accounts.length + ' ATIVAS' : (accounts[0].status ? accounts[0].status.toUpperCase() : 'ONLINE');
            }
          }

          if (accList && accounts.length > 0) {
            var botRam = (b.status && b.status.ramSysMB) ? b.status.ramSysMB.toFixed(1) + ' MB' : '< 10 MB';
            var botUptime = (b.status && b.status.uptime) ? b.status.uptime : '24/7 Ativo';

            var frontHtml = accounts.map(function(a) {
              var s = '<div class="acc-item">';
              if (isStream) {
                s += '<div class="acc-item-header"><span>' + (a.name || 'Conta') + '</span><span style="font-size: 0.72rem; font-weight: 700;">STREAM VIEWER</span></div>';
                if (a.currentArea) {
                  s += '<div class="stat-row" style="padding: 4px 8px; border-radius: 6px; margin-bottom: 4px; background: rgba(255,255,255,0.05); font-weight: 700;">📡 ' + a.currentArea + '</div>';
                }
                s += '<div class="stat-row"><span style="color: #facc15; font-weight: 800;">⭐ ' + (a.gold || 0).toLocaleString() + ' pontos</span><span style="color: #38bdf8; font-weight: 700;">🔴 ' + (a.kills || 0) + ' ' + ((a.kills === 1) ? 'canal ao vivo' : 'canais ao vivo') + '</span></div>';
              } else if (isService) {
                s += '<div class="acc-item-header"><span>' + (a.name || 'Serviço') + '</span><span style="font-size: 0.72rem; font-weight: 700;">' + (a.vocation || 'Service') + '</span></div>';
                if (a.currentArea) {
                  s += '<div class="stat-row" style="padding: 4px 8px; border-radius: 6px; margin-bottom: 4px; background: rgba(255,255,255,0.05); font-weight: 700;">⚙️ ' + a.currentArea + '</div>';
                }
                s += '<div class="stat-row"><span style="color: #4ade80; font-weight: 700;">🟢 ' + (a.status || 'Online') + '</span><span style="color: #38bdf8; font-weight: 600;">⚡ Operação Contínua</span></div>';
              } else {
                s += '<div class="acc-item-header"><span>' + (a.name || 'Conta') + ' (Nv. ' + (a.level || 1) + ')</span><span>' + (a.vocation || '') + '</span></div>';
                if (a.currentArea) {
                  s += '<div class="stat-row" style="background: rgba(255,255,255,0.05); padding: 4px 8px; border-radius: 6px; margin-bottom: 4px;">🗺️ ' + a.currentArea + '</div>';
                }
                if (a.maxHp) {
                  s += '<div class="stat-row"><span>❤️ HP</span><span style="color: #4ade80;">' + (a.hp || 0).toLocaleString() + ' / ' + (a.maxHp || 0).toLocaleString() + '</span></div>';
                }
                if (a.stamina) {
                  var stam = typeof a.stamina === 'number' ? a.stamina + '%' : a.stamina;
                  s += '<div class="stat-row"><span>⚡ Stamina</span><span style="color: #38bdf8; font-weight: 600;">' + stam + '</span></div>';
                }
                s += '<div class="stat-row"><span style="color: #facc15; font-weight: 800;">💰 ' + (a.gold || 0).toLocaleString() + ' gp</span><span style="color: #38bdf8;">⚔️ ' + (a.kills || 0).toLocaleString() + ' kills</span></div>';
              }
              s += '</div>';
              return s;
            }).join('');

            var backHtml = accounts.map(function(a) {
              var s = '<div class="acc-item">';
              if (isStream) {
                s += '<div class="acc-item-header"><span>⚡ Telemetria Drops</span><span class="perf-tag">AUTOMÁTICO</span></div>';
                s += '<div class="stat-row"><span>🎁 Coleta de Pontos</span><span style="color: #4ade80; font-weight: 700;">Ativo (Zero-Video WS)</span></div>';
                s += '<div class="stat-row"><span>🧠 Memória Bot (Go)</span><span style="color: #38bdf8; font-weight: 600;">' + botRam + '</span></div>';
                s += '<div class="stat-row"><span>⏱️ Uptime Contínuo</span><span style="color: #facc15;">' + formatUptime(a.durationSec, botUptime) + '</span></div>';
                s += '<div class="stat-row"><span>🛡️ Camuflagem IRC</span><span style="color: #a855f7; font-weight: 700;">100% Stealth Mode</span></div>';
              } else if (isService) {
                s += '<div class="acc-item-header"><span>⚡ Telemetria de Sistema</span><span class="perf-tag">ONLINE 24/7</span></div>';
                s += '<div class="stat-row"><span>🔄 Healthcheck</span><span style="color: #4ade80; font-weight: 700;">🟢 Heartbeat OK</span></div>';
                s += '<div class="stat-row"><span>🧠 Memória Isolada</span><span style="color: #38bdf8; font-weight: 600;">' + botRam + '</span></div>';
                s += '<div class="stat-row"><span>🛡️ Supervisor Resiliente</span><span style="color: #a855f7; font-weight: 700;">Goroutines com Recover</span></div>';
                s += '<div class="stat-row"><span>⏱️ Uptime</span><span style="color: #facc15;">' + formatUptime(a.durationSec, botUptime) + '</span></div>';
              } else {
                s += '<div class="acc-item-header"><span>⚡ ' + (a.name || 'Conta') + ' — Rendimento</span><span class="perf-tag">PROTEGIDO</span></div>';
                s += '<div class="stat-row"><span>📈 EXP / Hora</span><span style="color: #4ade80; font-weight: 700;">' + formatRate(a.expPerHour) + '</span></div>';
                s += '<div class="stat-row"><span>💰 Ouro / Hora</span><span style="color: #facc15; font-weight: 700;">' + formatRate(a.goldPerHour) + '</span></div>';
                s += '<div class="stat-row"><span>⏱️ Tempo de Caça</span><span style="color: #38bdf8;">' + formatUptime(a.durationSec, botUptime) + '</span></div>';
                s += '<div class="stat-row"><span>🧠 Memória Bot (Go)</span><span style="color: #94a3b8;">' + botRam + '</span></div>';
              }
              s += '</div>';
              return s;
            }).join('');

            accList.innerHTML = '<div class="card-face face-gameplay">' + frontHtml + '</div>' +
                                '<div class="card-face face-performance">' + backHtml + '</div>';
          }
        }
      } catch (e) {
        console.warn('Erro ao atualizar hub:', e);
      }
    }
    setInterval(refreshHub, 12000);
`
