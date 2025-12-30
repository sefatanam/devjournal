import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ThemeStore } from './theme.store';

@Component({
  selector: 'ui-theme-toggle',
  standalone: true,
  imports: [CommonModule],
  template: `
    <button 
      class="theme-toggle" 
      (click)="themeStore.toggle()"
      [attr.aria-label]="themeStore.isDark() ? 'Switch to light mode' : 'Switch to dark mode'"
      [title]="themeStore.isDark() ? 'Switch to light mode' : 'Switch to dark mode'"
    >
      <span class="theme-toggle__icon theme-toggle__icon--sun" [class.active]="themeStore.isLight()">
        <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <circle cx="12" cy="12" r="5"/>
          <line x1="12" y1="1" x2="12" y2="3"/>
          <line x1="12" y1="21" x2="12" y2="23"/>
          <line x1="4.22" y1="4.22" x2="5.64" y2="5.64"/>
          <line x1="18.36" y1="18.36" x2="19.78" y2="19.78"/>
          <line x1="1" y1="12" x2="3" y2="12"/>
          <line x1="21" y1="12" x2="23" y2="12"/>
          <line x1="4.22" y1="19.78" x2="5.64" y2="18.36"/>
          <line x1="18.36" y1="5.64" x2="19.78" y2="4.22"/>
        </svg>
      </span>
      <span class="theme-toggle__icon theme-toggle__icon--moon" [class.active]="themeStore.isDark()">
        <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/>
        </svg>
      </span>
    </button>
  `,
  styles: [`
    .theme-toggle {
      display: flex;
      align-items: center;
      justify-content: center;
      width: 40px;
      height: 40px;
      padding: 0;
      background: var(--color-surface);
      border: 1px solid var(--color-border);
      border-radius: var(--radius-md);
      cursor: pointer;
      position: relative;
      overflow: hidden;
      transition: var(--transition-all);
      
      &:hover {
        background: var(--color-surface-hover);
        border-color: var(--color-text-dim);
      }
      
      &:focus-visible {
        outline: 2px solid var(--color-primary);
        outline-offset: 2px;
      }
    }
    
    .theme-toggle__icon {
      position: absolute;
      display: flex;
      align-items: center;
      justify-content: center;
      color: var(--color-text-muted);
      transition: all 0.3s var(--ease-out-expo);
      
      &--sun {
        transform: translateY(-30px) rotate(-90deg);
        opacity: 0;
        
        &.active {
          transform: translateY(0) rotate(0);
          opacity: 1;
          color: var(--color-amber);
        }
      }
      
      &--moon {
        transform: translateY(0) rotate(0);
        opacity: 1;
        color: var(--color-cyan);
        
        &.active {
          transform: translateY(0) rotate(0);
          opacity: 1;
        }
      }
      
      &--sun.active ~ &--moon,
      &--moon:not(.active) {
        transform: translateY(30px) rotate(90deg);
        opacity: 0;
      }
    }
  `],
})
export class ThemeToggleComponent {
  readonly themeStore = inject(ThemeStore);
}
