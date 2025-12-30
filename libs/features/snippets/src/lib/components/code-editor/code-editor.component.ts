import {
  ChangeDetectionStrategy,
  Component,
  computed,
  input,
  output,
  signal,
} from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';

@Component({
  selector: 'lib-code-editor',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './code-editor.component.html',
  styleUrl: './code-editor.component.scss',
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class CodeEditorComponent {
  readonly code = input<string>('');
  readonly language = input<string>('typescript');
  readonly readonly = input<boolean>(false);
  readonly placeholder = input<string>('Enter your code here...');
  readonly minHeight = input<string>('300px');

  readonly codeChange = output<string>();

  protected readonly lineCount = computed(() => {
    const content = this.code();
    return content ? content.split('\n').length : 1;
  });

  protected readonly lineNumbers = computed(() => {
    return Array.from({ length: this.lineCount() }, (_, i) => i + 1);
  });

  protected onCodeChange(value: string): void {
    this.codeChange.emit(value);
  }

  protected onKeyDown(event: KeyboardEvent): void {
    const textarea = event.target as HTMLTextAreaElement;

    if (event.key === 'Tab') {
      event.preventDefault();
      const start = textarea.selectionStart;
      const end = textarea.selectionEnd;
      const value = textarea.value;

      const newValue = value.substring(0, start) + '  ' + value.substring(end);
      this.codeChange.emit(newValue);

      setTimeout(() => {
        textarea.selectionStart = textarea.selectionEnd = start + 2;
      }, 0);
    }
  }
}
