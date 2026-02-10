import type { Agent } from '../core/client.js';
import type { Form, FormResponse } from '../types/index.js';

export class FormHandler {
  private pendingResponses: Map<
    number,
    { resolve: (response: FormResponse) => void }
  > = new Map();

  constructor(private player: Agent) {
    this.setupListeners();
  }

  private setupListeners(): void {
    this.player.on('form', (form) => {
      // Auto-handle if a response is pending
      const pending = this.pendingResponses.get(form.id);
      if (pending) {
        // This is handled by waiting code
      }
    });
  }

  async waitForForm(options?: { timeout?: number }): Promise<Form> {
    const { timeout = 5000 } = options ?? {};

    const [form] = await this.player.waitFor('form', { timeout });
    return form;
  }

  respond(formId: number, response: FormResponse): void {
    this.player.respondToForm(formId, response);
  }

  close(formId: number): void {
    this.player.closeForm(formId);
  }
}
