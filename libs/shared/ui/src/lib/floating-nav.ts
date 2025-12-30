import { ChangeDetectionStrategy, Component, inject, PLATFORM_ID, computed } from '@angular/core';
import { isPlatformBrowser } from '@angular/common';
import { Router, RouterLink, RouterLinkActive } from '@angular/router';
import { AuthStore } from '@devjournal/feature-auth';
import { ThemeToggleComponent } from './theme/theme-toggle.component';
import { ThemeStore } from './theme/theme.store';

interface NavItem {
  path: string;
  label: string;
  icon: string;
}

@Component({
  selector: 'ui-floating-nav',
  imports: [RouterLink, RouterLinkActive, ThemeToggleComponent],
  templateUrl: './floating-nav.html',
  styleUrl: './floating-nav.scss',
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class FloatingNavComponent {
  private readonly authStore = inject(AuthStore);
  private readonly router = inject(Router);
  private readonly platformId = inject(PLATFORM_ID);
  protected readonly themeStore = inject(ThemeStore);

  protected readonly isBrowser = isPlatformBrowser(this.platformId);
  protected readonly isAuthenticated = computed(() => 
    this.isBrowser && this.authStore.isAuthenticated()
  );
  protected readonly userDisplayName = this.authStore.userDisplayName;

  protected readonly navItems: NavItem[] = [
    { path: '/dashboard', label: 'Home', icon: '⌂' },
    { path: '/journal', label: 'Journal', icon: '✎' },
    { path: '/snippets', label: 'Snippets', icon: '❮❯' },
    { path: '/progress', label: 'Progress', icon: '◈' },
    { path: '/chat', label: 'Chat', icon: '◉' },
  ];

  protected logout(): void {
    this.authStore.logout();
  }
}
