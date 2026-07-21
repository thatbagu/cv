// Animation Manager Class - handles smooth frame timing
class AnimationManager {
  constructor(callback, fps = 30) {
    this._animation = null;
    this.callback = callback;
    this.lastFrame = -1;
    this.frameTime = 1000 / fps;
  }

  start() {
    if (this._animation != null) return;
    this._animation = requestAnimationFrame(this.update);
  }

  pause() {
    if (this._animation == null) return;
    this.lastFrame = -1;
    cancelAnimationFrame(this._animation);
    this._animation = null;
  }

  update = (time) => {
    const { lastFrame } = this;
    let delta = time - lastFrame;
    if (this.lastFrame === -1) {
      this.lastFrame = time;
    } else {
      while (delta >= this.frameTime) {
        this.callback();
        delta -= this.frameTime;
        this.lastFrame += this.frameTime;
      }
    }
    this._animation = requestAnimationFrame(this.update);
  };
}

// ASCII Animation Component
class ASCIIAnimation {
  constructor(containerId, options = {}) {
    this.container = document.getElementById(containerId);
    if (!this.container) {
      console.error(`Container with id "${containerId}" not found`);
      return;
    }

    this.fps = options.fps || 24;
    this.colorOverlay = options.colorOverlay !== false;
    this.frameCount = options.frameCount || 60;
    this.framesPath = options.framesPath || "/frames";

    this.frames = [];
    this.currentFrame = 0;
    this.isLoading = true;
    this.animationManager = null;

    this.init();
  }

  async init() {
    // Create container structure
    this.container.innerHTML = `
      <div class="ascii-animation-wrapper">
        <div class="ascii-animation-loading">Loading animation...</div>
        <div class="ascii-animation-content" style="display: none;">
          <pre class="ascii-frame"></pre>
          ${this.colorOverlay ? '<div class="ascii-color-overlay"></div>' : ""}
        </div>
      </div>
    `;

    this.loadingEl = this.container.querySelector(".ascii-animation-loading");
    this.contentEl = this.container.querySelector(".ascii-animation-content");
    this.frameEl = this.container.querySelector(".ascii-frame");

    await this.loadFrames();
  }

  async loadFrames() {
    try {
      const frameFiles = Array.from(
        { length: this.frameCount },
        (_, i) => `frame_${String(i + 1).padStart(4, "0")}.txt`,
      );

      const framePromises = frameFiles.map(async (filename) => {
        const response = await fetch(`${this.framesPath}/${filename}`);
        if (!response.ok) {
          throw new Error(`Failed to fetch ${filename}: ${response.status}`);
        }
        return await response.text();
      });

      this.frames = await Promise.all(framePromises);
      this.isLoading = false;

      // Show content, hide loading
      this.loadingEl.style.display = "none";
      this.contentEl.style.display = "block";

      // Display first frame
      this.updateFrame();

      // Start animation
      this.startAnimation();
    } catch (error) {
      console.error("Failed to load ASCII frames:", error);
      this.loadingEl.textContent = "Failed to load animation";
    }
  }

  updateFrame() {
    if (this.frames.length === 0) return;
    this.frameEl.textContent = this.frames[this.currentFrame];
  }

  startAnimation() {
    // Check for reduced motion preference
    const reducedMotion = window.matchMedia(
      "(prefers-reduced-motion: reduce)",
    ).matches;
    if (reducedMotion) {
      return;
    }

    // Create animation manager
    this.animationManager = new AnimationManager(() => {
      this.currentFrame = (this.currentFrame + 1) % this.frames.length;
      this.updateFrame();
    }, this.fps);

    // Handle window focus/blur
    const handleFocus = () => this.animationManager.start();
    const handleBlur = () => this.animationManager.pause();

    window.addEventListener("focus", handleFocus);
    window.addEventListener("blur", handleBlur);

    // Start if window is visible
    if (document.visibilityState === "visible") {
      this.animationManager.start();
    }

    // Store cleanup function
    this.cleanup = () => {
      window.removeEventListener("focus", handleFocus);
      window.removeEventListener("blur", handleBlur);
      if (this.animationManager) {
        this.animationManager.pause();
      }
    };
  }

  destroy() {
    if (this.cleanup) {
      this.cleanup();
    }
  }
}

document.addEventListener("DOMContentLoaded", () => {
  const asciiAnim = new ASCIIAnimation("ascii-animation", {
    fps: 24,
    colorOverlay: true,
    frameCount: 171,
    framesPath: "/assets/index/animation/cat",
  });
});
