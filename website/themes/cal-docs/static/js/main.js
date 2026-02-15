// Theme toggle
(function () {
  var toggle = document.getElementById('theme-toggle');
  var html = document.documentElement;

  var saved = localStorage.getItem('theme');
  if (saved) {
    html.setAttribute('data-theme', saved);
  } else if (window.matchMedia('(prefers-color-scheme: dark)').matches) {
    html.setAttribute('data-theme', 'dark');
  }

  if (toggle) {
    toggle.addEventListener('click', function () {
      var current = html.getAttribute('data-theme');
      var next = current === 'dark' ? 'light' : 'dark';
      html.setAttribute('data-theme', next);
      localStorage.setItem('theme', next);
    });
  }
})();

// Mobile nav toggle
(function () {
  var btn = document.getElementById('nav-toggle');
  var links = document.getElementById('nav-links');
  if (btn && links) {
    btn.addEventListener('click', function () {
      links.classList.toggle('active');
    });
  }
})();

// Scroll reveal
(function () {
  var elements = document.querySelectorAll('.scroll-reveal');
  if (!elements.length) return;

  var observer = new IntersectionObserver(
    function (entries) {
      entries.forEach(function (entry) {
        if (entry.isIntersecting) {
          entry.target.classList.add('visible');
          observer.unobserve(entry.target);
        }
      });
    },
    { threshold: 0.12, rootMargin: '0px 0px -40px 0px' }
  );

  elements.forEach(function (el, i) {
    el.style.transitionDelay = (i % 3) * 0.08 + 's';
    observer.observe(el);
  });
})();

// Copy buttons (install section)
(function () {
  var copyIcon14 = '<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1"/></svg>';
  var checkIcon14 = '<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M20 6L9 17l-5-5"/></svg>';

  document.querySelectorAll('.install-copy-btn').forEach(function (btn) {
    btn.addEventListener('click', function () {
      var text = this.getAttribute('data-copy');
      navigator.clipboard.writeText(text).then(function () {
        btn.innerHTML = checkIcon14;
        btn.classList.add('copied');
        setTimeout(function () {
          btn.innerHTML = copyIcon14;
          btn.classList.remove('copied');
        }, 1500);
      });
    });
  });
})();

// Install tabs
(function () {
  var tabs = document.querySelectorAll('.install-tab');
  var panels = document.querySelectorAll('.install-panel');
  if (!tabs.length) return;

  tabs.forEach(function (tab) {
    tab.addEventListener('click', function () {
      var target = this.getAttribute('data-tab');
      tabs.forEach(function (t) {
        t.classList.remove('active');
        t.setAttribute('aria-selected', 'false');
      });
      panels.forEach(function (p) { p.classList.remove('active'); });
      this.classList.add('active');
      this.setAttribute('aria-selected', 'true');
      var panel = document.querySelector('[data-panel="' + target + '"]');
      if (panel) panel.classList.add('active');
    });
  });
})();

// Code block copy buttons (documentation pages)
(function () {
  var copyIcon = '<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1"/></svg>';
  var checkIcon = '<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M20 6L9 17l-5-5"/></svg>';

  document.querySelectorAll('.docs-body pre').forEach(function (pre) {
    var code = pre.querySelector('code');
    if (!code) return;

    var btn = document.createElement('button');
    btn.className = 'code-copy-btn';
    btn.setAttribute('aria-label', 'Copy code');
    btn.innerHTML = copyIcon;

    btn.addEventListener('click', function () {
      var text = code.textContent;
      navigator.clipboard.writeText(text).then(function () {
        btn.innerHTML = checkIcon;
        btn.classList.add('copied');
        setTimeout(function () {
          btn.innerHTML = copyIcon;
          btn.classList.remove('copied');
        }, 1500);
      });
    });

    pre.appendChild(btn);
  });
})();

// Terminal typing animation
(function () {
  var cmdEl = document.getElementById('typed-cmd');
  var outputEl = document.getElementById('terminal-output');
  if (!cmdEl || !outputEl) return;

  var demos = [
    {
      cmd: 'ical today',
      output:
        '#  Title              Calendar   Start        End\n1  Team standup        Work       9:00 AM      9:30 AM\n2  Design review       Work       2:00 PM      3:00 PM\n3  Gym                 Personal   6:00 PM      7:00 PM',
    },
    {
      cmd: 'ical add "Launch party" -s friday 5pm -e friday 8pm',
      output: 'Created event: Launch party (Work)',
    },
    {
      cmd: 'ical upcoming -d 3',
      output:
        '#  Title              Calendar   Date         Time\n1  Team standup        Work       Today        9:00 AM\n2  Design review       Work       Today        2:00 PM\n3  1:1 with Alex       Work       Tomorrow     10:00 AM\n4  Launch party        Work       Fri          5:00 PM',
    },
    {
      cmd: 'ical search "launch"',
      output:
        '#  Title              Calendar   Date         Time\n1  Launch party        Work       Fri Feb 14   5:00 PM',
    },
  ];

  var demoIdx = 0;
  var charIdx = 0;
  var typing = true;
  var pauseAfter = 2200;

  function typeNext() {
    var demo = demos[demoIdx];

    if (typing) {
      if (charIdx <= demo.cmd.length) {
        cmdEl.textContent = demo.cmd.slice(0, charIdx);
        charIdx++;
        setTimeout(typeNext, 32 + Math.random() * 28);
      } else {
        typing = false;
        setTimeout(function () {
          outputEl.textContent = demo.output;
          setTimeout(typeNext, pauseAfter);
        }, 300);
      }
    } else {
      outputEl.textContent = '';
      cmdEl.textContent = '';
      charIdx = 0;
      typing = true;
      demoIdx = (demoIdx + 1) % demos.length;
      setTimeout(typeNext, 400);
    }
  }

  setTimeout(typeNext, 800);
})();
