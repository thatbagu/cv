(async function () {
  const steppe = document.getElementById("steppe");
  if (!steppe) return;

  const isMobile = window.innerWidth < 768;

  const BASE = "/assets/index/animation/steppe/";
  const load = (path) =>
    fetch(BASE + path).then((r) => (r.ok ? r.text() : Promise.reject()));

  let roofs, bases, stoolArt, tableArt, fireplaceArt, horseFrames, horseIdle, horseEat,
      dogFrames, dogIdle;
  try {
    [roofs, bases, stoolArt, tableArt, fireplaceArt, horseFrames, horseIdle, horseEat,
     dogFrames, dogIdle] =
      await Promise.all([
        Promise.all([load("yurt_roof_left.txt"), load("yurt_roof_right.txt")]),
        Promise.all([
          load("yurt_base_open.txt"),
          load("yurt_base_open_clean.txt"),
          load("yurt_base_closed.txt"),
          load("yurt_base_closed_clean.txt"),
          load("yurt_base_light.txt"),
          load("yurt_base_light_clean.txt"),
        ]),
        load("stool.txt"),
        load("table.txt"),
        load("fireplace.txt"),
        Promise.all([
          load("horse_move_frame1.txt"),
          load("horse_move_frame2.txt"),
          load("horse_move_frame3.txt"),
        ]),
        load("horse_idle.txt"),
        load("horse_eat.txt"),
        Promise.all([load("dog_move_frame1.txt"), load("dog_move_frame2.txt")]),
        load("dog_idle.txt"),
      ]);
  } catch {
    return;
  }

  stoolArt     = stoolArt.trimEnd();
  tableArt     = tableArt.trimEnd();
  fireplaceArt = fireplaceArt.trimEnd();
  horseFrames  = horseFrames.map((f) => f.trimEnd());
  horseIdle    = horseIdle.trimEnd();
  horseEat     = horseEat.trimEnd();
  dogFrames    = dogFrames.map((f) => f.trimEnd());
  dogIdle      = dogIdle.trimEnd();

  // Fire animation frames — only top 2 rows change, bottom 3 are static
  const fireplaceLines = fireplaceArt.split("\n");
  const fireplaceBase  = fireplaceLines.slice(2).join("\n");
  const FIRE_TOPS = [
    "     (    \n    ).    ",
    "     )    \n    (.    ",
    "    ( )   \n    )(    ",
    "    |(    \n    .)    ",
    "     (    \n    .)    ",
    "      )   \n    (.    ",
  ];
  const FIRE_FRAMES = FIRE_TOPS.map((t) => t + "\n" + fireplaceBase);

  function pick(arr) {
    return arr[Math.floor(Math.random() * arr.length)];
  }

  const availableBases = isMobile ? bases.slice(0, 4) : bases;

  function randomYurt() {
    const roofIdx = Math.floor(Math.random() * roofs.length);
    const text = roofs[roofIdx].trimEnd() + "\n" + pick(availableBases).trimEnd();
    return { text, roofIdx };
  }

  const BOTTOM_MAX = steppe.offsetHeight * 0.85;
  const RANGE      = 5 * 5;

  function sortByDepth() {
    const children = Array.from(steppe.children);
    children.sort((a, b) => parseFloat(b.dataset.depth) - parseFloat(a.dataset.depth));
    children.forEach((el) => steppe.appendChild(el));
  }

  function makeEl(tag, cls, text, x, bottom) {
    const el = document.createElement(tag);
    el.className      = cls;
    el.style.position = "absolute";
    el.style.left     = typeof x === "string" ? x : x + "%";
    el.style.bottom   = bottom + "px";
    el.dataset.depth  = bottom;
    if (text != null) el.textContent = text;
    steppe.appendChild(el);
    return el;
  }

  // Overlap guard — must be defined before first use
  const occupied = [];
  function noOverlap(xPct, bottom, xThresh = 10, yThresh = 15) {
    return occupied.every(
      (o) => Math.abs(xPct - o.xPct) > xThresh || Math.abs(bottom - o.bottom) > yThresh
    );
  }

  // Yurts + stools
  const yurtInfos = [];
  const yurtCount = isMobile
    ? 1 + Math.floor(Math.random() * 2)   // 1–2 on mobile
    : 1 + Math.floor(Math.random() * 4);  // 1–4 on desktop

  Array.from({ length: yurtCount }, (_, i) => {
    const min = i * (100 / yurtCount) + 4;
    const max = (i + 1) * (100 / yurtCount) - 4;
    return min + Math.random() * (max - min);
  }).forEach((x) => {
    const bottom = Math.random() * RANGE;
    const { text, roofIdx } = randomYurt();
    const el = makeEl("pre", "steppe-yurt", text, x, bottom);
    yurtInfos.push({ el, bottom, roofIdx, xPct: x });
    occupied.push({ xPct: x, bottom });

    const maxStools = isMobile ? 2 : 4;
    const stoolCount = Math.floor(Math.random() * maxStools);
    for (let s = 0; s < stoolCount; s++) {
      const side = Math.random() < 0.5 ? -1 : 1;
      const sx = x + side * (5 + Math.random() * 6);
      if (!noOverlap(sx, bottom, 4, 10)) continue;
      makeEl("pre", "steppe-yurt", stoolArt, sx, bottom);
      occupied.push({ xPct: sx, bottom });
      // occasionally place a table next to this stool
      if (Math.random() < 0.35) {
        const tx = sx + side * (3 + Math.random() * 3);
        if (noOverlap(tx, bottom, 5, 10)) {
          makeEl("pre", "steppe-yurt", tableArt, tx, bottom);
          occupied.push({ xPct: tx, bottom });
        }
      }
    }
  });

  // Fireplaces — check x-only against yurts (yurts always overlap in y given small RANGE)
  function clearOfYurts(xPct) {
    return yurtInfos.every((y) => Math.abs(xPct - y.xPct) > 15);
  }

  const fireplaceAnims = [];
  const maxFires = isMobile ? 2 : Math.min(5, yurtCount + 1);
  const fireCount = 1 + Math.floor(Math.random() * maxFires);
  for (let f = 0; f < fireCount; f++) {
    let xPct, bottom, tries = 0;
    do {
      xPct   = 4 + Math.random() * 92;
      bottom = Math.random() * RANGE;
      tries++;
    } while ((!noOverlap(xPct, bottom) || !clearOfYurts(xPct)) && tries < 40);
    if (tries < 40) {
      const el = makeEl("pre", "steppe-yurt", FIRE_FRAMES[0], xPct, bottom);
      fireplaceAnims.push({
        el,
        frameIdx: Math.floor(Math.random() * FIRE_FRAMES.length),
        timer:    Math.random() * 150,
        interval: 100 + Math.random() * 80,
      });
      occupied.push({ xPct, bottom });
    }
  }

  // Grass
  const GRASS = ["W", "w", '"', "'", "^"];
  const grassCount = isMobile ? 20 : 36 + Math.floor(Math.random() * 54);
  for (let g = 0; g < grassCount; g++) {
    if (Math.random() > 0.4) continue;
    const xPct = Math.random() * 96 + 2;
    const b    = Math.random() * steppe.offsetHeight * 0.85;
    if (!noOverlap(xPct, b)) continue;
    const span = makeEl("span", "steppe-yurt", pick(GRASS), xPct + "%", b);
    span.style.transform = "translateX(-50%)";
  }

  sortByDepth();

  // Smoke
  const CHIMNEY_COL = [23, 7];
  const SMOKE_CHARS = ["~", ",", "~", ".", "~", "'"];
  const CHAR_W      = 3;
  const YURT_HALF_W = 45;
  const CHIMNEY_ROW = 40;

  const windX = (Math.random() < 0.5 ? 1 : -1) * (6 + Math.random() * 10);
  const windY = 20 + Math.random() * 12;

  class SmokeSystem {
    constructor(x, bottom) {
      this.x         = x;
      this.bottom    = bottom;
      this.particles = [];
      this.interval  = 180 + Math.random() * 120; // 180–300 ms → dense but not overwhelming
      this.timer     = this.interval;
    }

    spawn() {
      const el = document.createElement("span");
      el.className     = "steppe-smoke";
      el.textContent   = pick(SMOKE_CHARS);
      el.dataset.depth = -9999;
      steppe.appendChild(el);
      const px = this.x + (Math.random() - 0.5) * 4;
      const pb = this.bottom;
      el.style.left    = px + "px";
      el.style.bottom  = pb + "px";
      el.style.opacity = "1";
      this.particles.push({
        el,
        x: px, bottom: pb,
        vy: windY + (Math.random() - 0.5) * 6,
        vx: windX + (Math.random() - 0.5) * 4,
        opacity: 1,
        decay: 0.15 + Math.random() * 0.12,
      });
    }

    tick(dt) {
      this.timer += dt;
      if (this.timer >= this.interval) {
        this.timer = 0;
        this.spawn();
      }
      this.particles = this.particles.filter((p) => {
        p.x      += p.vx * (dt / 1000);
        p.bottom += p.vy * (dt / 1000);
        p.opacity -= p.decay * (dt / 1000);
        p.el.style.left    = p.x + "px";
        p.el.style.bottom  = p.bottom + "px";
        p.el.style.opacity = Math.max(0, p.opacity);
        if (p.opacity <= 0) { p.el.remove(); return false; }
        return true;
      });
    }
  }

  const steppeW = steppe.offsetWidth;
  const smokeSystems = yurtInfos.flatMap(({ bottom: yb, roofIdx, xPct }) => {
    if (Math.random() < 0.35) return [];
    const col           = CHIMNEY_COL[roofIdx];
    const x             = (xPct / 100) * steppeW - YURT_HALF_W + col * CHAR_W;
    const chimneyBottom = yb + CHIMNEY_ROW;
    return [new SmokeSystem(x, chimneyBottom)];
  });

  // Horses
  class Horse {
    constructor() {
      const angle     = Math.random() * Math.PI * 2;
      this.speed      = 20 + Math.random() * 20;
      this.vx         = Math.cos(angle);
      this.vy         = Math.sin(angle);
      this.x          = Math.random() * steppe.offsetWidth;
      this.bottom     = Math.random() * BOTTOM_MAX;
      this.state      = "walk";
      this.frameIdx   = 0;
      this.frameTimer = 0;
      this.stateTimer = 2000 + Math.random() * 3000;
      this.el         = makeEl("pre", "steppe-horse", null, this.x + "px", this.bottom);
      this.render();
    }

    pickDirection() {
      const angle = Math.random() * Math.PI * 2;
      this.vx = Math.cos(angle);
      this.vy = Math.sin(angle);
    }

    render() {
      this.el.textContent =
        this.state === "idle" ? horseIdle
        : this.state === "eat"  ? horseEat
        : horseFrames[this.frameIdx];
      this.el.style.left      = this.x + "px";
      this.el.style.bottom    = this.bottom + "px";
      this.el.dataset.depth   = this.bottom;
      this.el.style.transform = this.vx < 0 ? "scaleX(-1)" : "scaleX(1)";
    }

    tick(dt) {
      this.frameTimer += dt;
      this.stateTimer -= dt;

      // Dog repulsion — steer away when dog is within detection ellipse
      if (dogs.length > 0) {
        const dog = dogs[0];
        const dx = this.x - dog.x;
        const dy = this.bottom - dog.bottom;
        // Normalise each axis by its detection radius so the check is elliptical
        if ((dx / 130) ** 2 + (dy / 18) ** 2 < 1) {
          // Wake up idle/eating horses
          if (this.state !== "walk") {
            this.state = "walk";
            this.stateTimer = 1500 + Math.random() * 2000;
          }
          // Blend current direction toward the away-vector
          const len = Math.sqrt(dx * dx + dy * dy) || 1;
          const ax = dx / len;
          const ay = dy / len;
          this.vx = this.vx * 0.7 + ax * 0.3;
          this.vy = this.vy * 0.7 + ay * 0.3;
          const vlen = Math.sqrt(this.vx ** 2 + this.vy ** 2) || 1;
          this.vx /= vlen;
          this.vy /= vlen;
          this.stateTimer = Math.min(this.stateTimer, 800); // keep fleeing
        }
      }

      if (this.stateTimer <= 0) {
        this.stateTimer = 2000 + Math.random() * 4000;
        const r = Math.random();
        this.state = r < 0.55 ? "walk" : r < 0.8 ? "idle" : "eat";
        if (this.state === "walk") this.pickDirection();
      }

      if (this.state === "walk") {
        const s = this.speed * (dt / 1000);
        this.x      += this.vx * s;
        this.bottom += this.vy * s;

        const W = steppe.offsetWidth;
        if (this.x > W + 80)      this.x = -80;
        if (this.x < -80)         this.x = W + 80;
        if (this.bottom < 0)          { this.bottom = 0;          this.vy *= -1; }
        if (this.bottom > BOTTOM_MAX) { this.bottom = BOTTOM_MAX; this.vy *= -1; }

        if (this.frameTimer >= 200) {
          this.frameTimer = 0;
          this.frameIdx = (this.frameIdx + 1) % horseFrames.length;
        }
      }

      this.render();
    }
  }

  const horseCount = isMobile
    ? 1 + Math.floor(Math.random() * 3)
    : 1 + Math.floor(Math.random() * 7);
  const horses = Array.from({ length: horseCount }, () => new Horse());

  // Dog — 0 or 1, walks/idles like a horse but smaller and quicker
  class Dog {
    constructor() {
      const angle     = Math.random() * Math.PI * 2;
      this.speed      = 30 + Math.random() * 25;
      this.vx         = Math.cos(angle);
      this.vy         = Math.sin(angle);
      this.x          = Math.random() * steppe.offsetWidth;
      this.bottom     = Math.random() * BOTTOM_MAX;
      this.state      = "walk";
      this.frameIdx   = 0;
      this.frameTimer = 0;
      this.stateTimer = 1500 + Math.random() * 2500;
      this.el         = makeEl("pre", "steppe-horse", null, this.x + "px", this.bottom);
      this.render();
    }

    pickDirection() {
      const angle = Math.random() * Math.PI * 2;
      this.vx = Math.cos(angle);
      this.vy = Math.sin(angle);
    }

    render() {
      this.el.textContent   = this.state === "idle" ? dogIdle : dogFrames[this.frameIdx];
      this.el.style.left    = this.x + "px";
      this.el.style.bottom  = this.bottom + "px";
      this.el.dataset.depth = this.bottom;
      this.el.style.transform = this.vx < 0 ? "scaleX(1)" : "scaleX(-1)";
    }

    tick(dt) {
      this.frameTimer += dt;
      this.stateTimer -= dt;

      if (this.stateTimer <= 0) {
        this.stateTimer = 1500 + Math.random() * 3000;
        this.state = Math.random() < 0.65 ? "walk" : "idle";
        if (this.state === "walk") this.pickDirection();
      }

      if (this.state === "walk") {
        const s = this.speed * (dt / 1000);
        this.x      += this.vx * s;
        this.bottom += this.vy * s;

        const W = steppe.offsetWidth;
        if (this.x > W + 40)      this.x = -40;
        if (this.x < -40)         this.x = W + 40;
        if (this.bottom < 0)          { this.bottom = 0;          this.vy *= -1; }
        if (this.bottom > BOTTOM_MAX) { this.bottom = BOTTOM_MAX; this.vy *= -1; }

        if (this.frameTimer >= 150) {
          this.frameTimer = 0;
          this.frameIdx = (this.frameIdx + 1) % dogFrames.length;
        }
      }

      this.render();
    }
  }

  const dogs = Math.random() < 0.5 ? [new Dog()] : [];

  let lastTime = null;
  let animId   = null;

  function updateAnimalVisibility(animal) {
    const hr = animal.el.getBoundingClientRect();
    const hidden = yurtInfos.some(({ el, bottom }) => {
      if (animal.bottom <= bottom) return false;
      const yr = el.getBoundingClientRect();
      return hr.left < yr.right && hr.right > yr.left;
    });
    animal.el.style.visibility = hidden ? "hidden" : "visible";
  }

  function animate(now) {
    const dt = lastTime === null ? 0 : now - lastTime;
    lastTime = now;

    horses.forEach((h) => { h.tick(dt); updateAnimalVisibility(h); });
    dogs.forEach((d) => { d.tick(dt); updateAnimalVisibility(d); });
    smokeSystems.forEach((s) => s.tick(dt));

    fireplaceAnims.forEach((f) => {
      f.timer += dt;
      if (f.timer >= f.interval) {
        f.timer    = 0;
        f.interval = 90 + Math.random() * 110;
        f.frameIdx = (f.frameIdx + 1) % FIRE_FRAMES.length;
        f.el.textContent = FIRE_FRAMES[f.frameIdx];
      }
    });

    sortByDepth();
    animId = requestAnimationFrame(animate);
  }

  document.addEventListener("visibilitychange", () => {
    if (document.hidden) {
      cancelAnimationFrame(animId);
      animId   = null;
      lastTime = null;
    } else if (!animId) {
      animId = requestAnimationFrame(animate);
    }
  });

  animId = requestAnimationFrame(animate);
})();
