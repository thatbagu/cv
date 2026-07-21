(async function () {
  const steppe = document.getElementById("steppe");
  if (!steppe) return;

  const BASE = "/assets/index/animation/steppe/";
  const load = (path) =>
    fetch(BASE + path).then((r) => (r.ok ? r.text() : Promise.reject()));

  let roofs, bases, stoolArt, fireplaceArt, horseFrames, horseIdle, horseEat;
  try {
    [roofs, bases, stoolArt, fireplaceArt, horseFrames, horseIdle, horseEat] =
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
        load("fireplace.txt"),
        Promise.all([
          load("horse_move_frame1.txt"),
          load("horse_move_frame2.txt"),
          load("horse_move_frame3.txt"),
        ]),
        load("horse_idle.txt"),
        load("horse_eat.txt"),
      ]);
  } catch {
    return;
  }

  stoolArt = stoolArt.trimEnd();
  fireplaceArt = fireplaceArt.trimEnd();
  horseFrames = horseFrames.map((f) => f.trimEnd());
  horseIdle = horseIdle.trimEnd();
  horseEat = horseEat.trimEnd();

  function pick(arr) {
    return arr[Math.floor(Math.random() * arr.length)];
  }

  function randomYurt() {
    const roofIdx = Math.floor(Math.random() * roofs.length); // 0=left, 1=right
    const text = roofs[roofIdx].trimEnd() + "\n" + pick(bases).trimEnd();
    return { text, roofIdx };
  }

  const BOTTOM_MAX = steppe.offsetHeight * 0.85;
  const RANGE = 5 * 5; // 5 char heights at 5px

  // Painter's algorithm: sort all children by depth (high bottom = far = paint first)
  function sortByDepth() {
    const children = Array.from(steppe.children);
    children.sort(
      (a, b) => parseFloat(b.dataset.depth) - parseFloat(a.dataset.depth),
    );
    children.forEach((el) => steppe.appendChild(el));
  }

  function makeEl(tag, cls, text, x, bottom) {
    const el = document.createElement(tag);
    el.className = cls;
    el.style.position = "absolute";
    el.style.left = typeof x === "string" ? x : x + "%";
    el.style.bottom = bottom + "px";
    el.dataset.depth = bottom;
    if (text != null) el.textContent = text;
    steppe.appendChild(el);
    return el;
  }

  // Yurts + stools
  const yurtInfos = []; // { el, bottom, roofIdx, xPct }
  const occupied = []; // { xPct, bottom } — all static positions for collision avoidance

  const yurtCount = 1 + Math.floor(Math.random() * 4);
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
    const stoolCount = Math.floor(Math.random() * 4);
    for (let s = 0; s < stoolCount; s++) {
      const side = Math.random() < 0.5 ? -1 : 1;
      const sx = x + side * (5 + Math.random() * 6);
      makeEl("pre", "steppe-yurt", stoolArt, sx, bottom);
      occupied.push({ xPct: sx, bottom });
    }
  });

  // Fireplaces — 1–5, placed without overlapping existing elements
  function noOverlap(xPct, bottom) {
    return occupied.every(
      (o) => Math.abs(xPct - o.xPct) > 10 || Math.abs(bottom - o.bottom) > 15,
    );
  }

  const fireCount = 1 + Math.floor(Math.random() * Math.min(5, yurtCount + 1));
  for (let f = 0; f < fireCount; f++) {
    let xPct,
      bottom,
      tries = 0;
    do {
      xPct = 4 + Math.random() * 92;
      bottom = Math.random() * RANGE;
      tries++;
    } while (!noOverlap(xPct, bottom) && tries < 30);
    if (tries < 30) {
      makeEl("pre", "steppe-yurt", fireplaceArt, xPct, bottom);
      occupied.push({ xPct, bottom });
    }
  }

  // Grass — placed after yurts/fireplaces so it avoids them
  const GRASS = ["W", "w", '"', "'", "^"];
  const grassCount = 36 + Math.floor(Math.random() * 54);
  for (let g = 0; g < grassCount; g++) {
    if (Math.random() > 0.4) continue;
    const xPct = Math.random() * 96 + 2;
    const b = Math.random() * steppe.offsetHeight * 0.85;
    if (!noOverlap(xPct, b)) continue;
    const span = makeEl("span", "steppe-yurt", pick(GRASS), xPct + "%", b);
    span.style.transform = "translateX(-50%)";
  }

  // Sort static elements once
  sortByDepth();

  // Smoke — chimney positions derived from known yurt data (no getBoundingClientRect needed)
  // Yurt art is ~30 chars wide at 3px/char = 90px; half = 45px from center to left edge
  // Chimney col 23 (left roof) / col 7 (right roof), row index 1 from top
  // Yurt has 10 lines (6 roof + 4 base) × 5px = 50px tall
  // chimneyBottom from steppe bottom = yurtBottom + (10 - 1 - 1) * 5 = yurtBottom + 40
  const CHIMNEY_COL = [23, 7];
  const SMOKE_CHARS = ["~", ",", "~", ".", "~", "'"];
  const CHAR_W = 3;
  const YURT_HALF_W = 45;
  const CHIMNEY_ROW = 40; // px above yurt bottom

  // Single wind direction for the whole scene — smoke drifts diagonally
  const windX = (Math.random() < 0.5 ? 1 : -1) * (6 + Math.random() * 10);
  const windY = 20 + Math.random() * 12;

  class SmokeSystem {
    constructor(x, bottom) {
      this.x = x;
      this.bottom = bottom;
      this.particles = [];
      this.interval = 25 + Math.random() * 350;
      this.timer = this.interval; // spawn on first tick
    }

    spawn() {
      const el = document.createElement("span");
      el.className = "steppe-smoke";
      el.textContent = pick(SMOKE_CHARS);
      el.dataset.depth = -9999;
      steppe.appendChild(el);
      const px = this.x + (Math.random() - 0.5) * 4;
      const pb = this.bottom;
      el.style.left = px + "px";
      el.style.bottom = pb + "px";
      el.style.opacity = "1";
      this.particles.push({
        el,
        x: px,
        bottom: pb,
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
        p.x += p.vx * (dt / 1000);
        p.bottom += p.vy * (dt / 1000);
        p.opacity -= p.decay * (dt / 1000);
        p.el.style.left = p.x + "px";
        p.el.style.bottom = p.bottom + "px";
        p.el.style.opacity = Math.max(0, p.opacity);
        if (p.opacity <= 0) {
          p.el.remove();
          return false;
        }
        return true;
      });
    }
  }

  const steppeW = steppe.offsetWidth;
  const smokeSystems = yurtInfos.flatMap(
    ({ bottom: yurtBottom, roofIdx, xPct }) => {
      if (Math.random() < 0.35) return [];
      const col = CHIMNEY_COL[roofIdx];
      const x = (xPct / 100) * steppeW - YURT_HALF_W + col * CHAR_W;
      const chimneyBottom = yurtBottom + CHIMNEY_ROW;
      return [new SmokeSystem(x, chimneyBottom)];
    },
  );

  // Horses
  class Horse {
    constructor() {
      const angle = Math.random() * Math.PI * 2;
      this.speed = 20 + Math.random() * 20;
      this.vx = Math.cos(angle);
      this.vy = Math.sin(angle);
      this.x = Math.random() * steppe.offsetWidth;
      this.bottom = Math.random() * BOTTOM_MAX;
      this.state = "walk";
      this.frameIdx = 0;
      this.frameTimer = 0;
      this.stateTimer = 2000 + Math.random() * 3000;
      this.el = makeEl("pre", "steppe-horse", null, this.x + "px", this.bottom);
      this.render();
    }

    pickDirection() {
      const angle = Math.random() * Math.PI * 2;
      this.vx = Math.cos(angle);
      this.vy = Math.sin(angle);
    }

    render() {
      this.el.textContent =
        this.state === "idle"
          ? horseIdle
          : this.state === "eat"
            ? horseEat
            : horseFrames[this.frameIdx];
      this.el.style.left = this.x + "px";
      this.el.style.bottom = this.bottom + "px";
      this.el.dataset.depth = this.bottom;
      this.el.style.transform = this.vx < 0 ? "scaleX(-1)" : "scaleX(1)";
    }

    tick(dt) {
      this.frameTimer += dt;
      this.stateTimer -= dt;

      if (this.stateTimer <= 0) {
        this.stateTimer = 2000 + Math.random() * 4000;
        const r = Math.random();
        this.state = r < 0.55 ? "walk" : r < 0.8 ? "idle" : "eat";
        if (this.state === "walk") this.pickDirection();
      }

      if (this.state === "walk") {
        const s = this.speed * (dt / 1000);
        this.x += this.vx * s;
        this.bottom += this.vy * s;

        const W = steppe.offsetWidth;
        if (this.x > W + 80) this.x = -80;
        if (this.x < -80) this.x = W + 80;
        if (this.bottom < 0) {
          this.bottom = 0;
          this.vy *= -1;
        }
        if (this.bottom > BOTTOM_MAX) {
          this.bottom = BOTTOM_MAX;
          this.vy *= -1;
        }

        if (this.frameTimer >= 200) {
          this.frameTimer = 0;
          this.frameIdx = (this.frameIdx + 1) % horseFrames.length;
        }
      }

      this.render();
    }
  }

  const horses = Array.from(
    { length: 1 + Math.floor(Math.random() * 7) },
    () => new Horse(),
  );

  let lastTime = null;
  let animId = null;

  function updateHorseVisibility(horse) {
    const hr = horse.el.getBoundingClientRect();
    const hidden = yurtInfos.some(({ el, bottom }) => {
      if (horse.bottom <= bottom) return false; // horse is in front of this yurt
      const yr = el.getBoundingClientRect();
      return hr.left < yr.right && hr.right > yr.left; // horizontal overlap
    });
    horse.el.style.visibility = hidden ? "hidden" : "visible";
  }

  function animate(now) {
    const dt = lastTime === null ? 0 : now - lastTime;
    lastTime = now;
    horses.forEach((h) => {
      h.tick(dt);
      updateHorseVisibility(h);
    });
    smokeSystems.forEach((s) => s.tick(dt));
    sortByDepth();
    animId = requestAnimationFrame(animate);
  }

  document.addEventListener("visibilitychange", () => {
    if (document.hidden) {
      cancelAnimationFrame(animId);
      animId = null;
      lastTime = null;
    } else if (!animId) animId = requestAnimationFrame(animate);
  });

  animId = requestAnimationFrame(animate);
})();
