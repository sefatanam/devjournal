import { Component } from '@angular/core';
import { RouterOutlet } from '@angular/router';
import { FloatingNavComponent } from '@devjournal/shared-ui';

@Component({
  selector: 'app-root',
  imports: [RouterOutlet, FloatingNavComponent],
  template: `
    <router-outlet />
    <ui-floating-nav />
  `,
  styles: [
    `
      :host {
        display: block;
      }
    `,
  ],
})
export class App {}
