import { Component, signal, OnInit, OnDestroy, PLATFORM_ID, inject } from '@angular/core';
import { CommonModule, isPlatformBrowser } from '@angular/common';
import { RouterLink } from '@angular/router';
import { ThemeToggleComponent, ThemeStore } from '@devjournal/shared-ui';

@Component({
  selector: 'app-landing',
  standalone: true,
  imports: [CommonModule, RouterLink, ThemeToggleComponent],
  templateUrl: './landing.component.html',
  styleUrl: './landing.component.scss',
})
export class LandingComponent implements OnInit, OnDestroy {
  private platformId = inject(PLATFORM_ID);
  protected themeStore = inject(ThemeStore);
  
  typedText = signal('');
  cursorVisible = signal(true);
  activeFeature = signal(0);
  streakCount = signal(0);
  
  private readonly heroTexts = [
    'document your journey_',
    'capture every insight_',
    'track your growth_',
    'build in public_',
  ];
  private currentTextIndex = 0;
  private charIndex = 0;
  private isDeleting = false;
  private typingInterval: ReturnType<typeof setInterval> | null = null;
  private cursorInterval: ReturnType<typeof setInterval> | null = null;
  private featureInterval: ReturnType<typeof setInterval> | null = null;
  private streakInterval: ReturnType<typeof setInterval> | null = null;

  readonly features = [
    {
      icon: 'journal',
      title: 'Dev Journal',
      description: 'Chronicle your coding adventures with rich markdown entries. Tag, search, and reflect on your daily progress.',
      command: '$ devjournal log --today',
      accent: 'cyan',
    },
    {
      icon: 'snippets',
      title: 'Snippet Vault',
      description: 'Your personal library of battle-tested code. Syntax-highlighted, searchable, always accessible.',
      command: '$ devjournal snippet --save',
      accent: 'amber',
    },
    {
      icon: 'progress',
      title: 'Streak Tracker',
      description: 'Build unbreakable habits. Visualize your consistency with heatmaps and streak counters.',
      command: '$ devjournal streak --status',
      accent: 'emerald',
    },
    {
      icon: 'chat',
      title: 'Study Rooms',
      description: 'Real-time collaboration with fellow devs. Share knowledge, solve problems together.',
      command: '$ devjournal chat --join',
      accent: 'coral',
    },
  ];

  readonly stats = [
    { value: '10K+', label: 'Developers' },
    { value: '500K+', label: 'Journal Entries' },
    { value: '1M+', label: 'Snippets Saved' },
    { value: '99.9%', label: 'Uptime' },
  ];

  // Pre-generated heatmap data to avoid SSR hydration mismatch
  readonly heatmapData = [
    [0.9, 0.7, 0.8, 0.6, 0.9, 0.5, 0.8],
    [0.6, 0.8, 0.9, 0.7, 0.5, 0.9, 0.7],
    [0.8, 0.5, 0.7, 0.9, 0.8, 0.6, 0.9],
    [0.7, 0.9, 0.6, 0.8, 0.7, 0.9, 0.5],
    [0.5, 0.7, 0.9, 0.6, 0.9, 0.8, 0.7],
    [0.9, 0.6, 0.8, 0.7, 0.5, 0.7, 0.9],
    [0.8, 0.9, 0.5, 0.9, 0.8, 0.6, 0.8],
  ];

  ngOnInit(): void {
    if (isPlatformBrowser(this.platformId)) {
      this.startTyping();
      this.startCursorBlink();
      this.startFeatureRotation();
      this.animateStreak();
    }
  }

  ngOnDestroy(): void {
    if (this.typingInterval) clearInterval(this.typingInterval);
    if (this.cursorInterval) clearInterval(this.cursorInterval);
    if (this.featureInterval) clearInterval(this.featureInterval);
    if (this.streakInterval) clearInterval(this.streakInterval);
  }

  private startTyping(): void {
    this.typingInterval = setInterval(() => {
      const currentText = this.heroTexts[this.currentTextIndex];
      
      if (!this.isDeleting) {
        this.typedText.set(currentText.substring(0, this.charIndex + 1));
        this.charIndex++;
        
        if (this.charIndex === currentText.length) {
          this.isDeleting = true;
          setTimeout(() => {}, 2000);
        }
      } else {
        this.typedText.set(currentText.substring(0, this.charIndex - 1));
        this.charIndex--;
        
        if (this.charIndex === 0) {
          this.isDeleting = false;
          this.currentTextIndex = (this.currentTextIndex + 1) % this.heroTexts.length;
        }
      }
    }, this.isDeleting ? 50 : 100);
  }

  private startCursorBlink(): void {
    this.cursorInterval = setInterval(() => {
      this.cursorVisible.update(v => !v);
    }, 530);
  }

  private startFeatureRotation(): void {
    this.featureInterval = setInterval(() => {
      this.activeFeature.update(v => (v + 1) % this.features.length);
    }, 4000);
  }

  private animateStreak(): void {
    const target = 127;
    const duration = 2000;
    const steps = 60;
    const increment = target / steps;
    let current = 0;
    
    this.streakInterval = setInterval(() => {
      current += increment;
      if (current >= target) {
        this.streakCount.set(target);
        if (this.streakInterval) clearInterval(this.streakInterval);
      } else {
        this.streakCount.set(Math.floor(current));
      }
    }, duration / steps);
  }

  setActiveFeature(index: number): void {
    this.activeFeature.set(index);
  }
}
