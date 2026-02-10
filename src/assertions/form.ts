import type { Agent } from '../core/client.js';
import type {
  Form,
  ModalForm,
  ActionForm,
  CustomForm,
  FormResponse,
} from '../types/index.js';
import { AssertionError } from './index.js';

export class FormAssertion {
  constructor(private player: Agent) {}

  async toReceive<T extends Form = Form>(options?: {
    timeout?: number;
    type?: T['type'];
  }): Promise<FormAssertionChain<T>> {
    const { timeout = 5000, type } = options ?? {};

    const filter = (form: Form): boolean => {
      if (type && form.type !== type) return false;
      return true;
    };

    try {
      const [form] = await this.player.waitFor('form', { timeout, filter });

      if (type === 'modal') {
        return new ModalFormAssertion(
          form as ModalForm,
          this.player
        ) as unknown as FormAssertionChain<T>;
      } else if (type === 'action') {
        return new ActionFormAssertion(
          form as ActionForm,
          this.player
        ) as unknown as FormAssertionChain<T>;
      } else if (type === 'form') {
        return new CustomFormAssertion(
          form as CustomForm,
          this.player
        ) as unknown as FormAssertionChain<T>;
      }

      // Auto-detect type
      if (form.type === 'modal') {
        return new ModalFormAssertion(
          form as ModalForm,
          this.player
        ) as unknown as FormAssertionChain<T>;
      } else if (form.type === 'action') {
        return new ActionFormAssertion(
          form as ActionForm,
          this.player
        ) as unknown as FormAssertionChain<T>;
      } else {
        return new CustomFormAssertion(
          form as CustomForm,
          this.player
        ) as unknown as FormAssertionChain<T>;
      }
    } catch {
      throw new AssertionError(
        `Timeout waiting for form${type ? ` of type "${type}"` : ''}`,
        type,
        undefined
      );
    }
  }

  async toReceiveModal(options?: { timeout?: number }): Promise<ModalFormAssertion> {
    return this.toReceive({ ...options, type: 'modal' }) as Promise<ModalFormAssertion>;
  }

  async toReceiveAction(options?: { timeout?: number }): Promise<ActionFormAssertion> {
    return this.toReceive({ ...options, type: 'action' }) as Promise<ActionFormAssertion>;
  }

  async toReceiveCustom(options?: { timeout?: number }): Promise<CustomFormAssertion> {
    return this.toReceive({ ...options, type: 'form' }) as Promise<CustomFormAssertion>;
  }
}

type FormAssertionChain<T extends Form> = T extends ModalForm
  ? ModalFormAssertion
  : T extends ActionForm
    ? ActionFormAssertion
    : CustomFormAssertion;

export class ModalFormAssertion {
  constructor(
    private form: ModalForm,
    private player: Agent
  ) {}

  toHaveTitle(expected: string | RegExp): this {
    const matches =
      typeof expected === 'string'
        ? this.form.title === expected
        : expected.test(this.form.title);

    if (!matches) {
      throw new AssertionError(
        `Expected form title to match ${expected}, but was "${this.form.title}"`,
        expected,
        this.form.title
      );
    }
    return this;
  }

  toHaveContent(expected: string | RegExp): this {
    const matches =
      typeof expected === 'string'
        ? this.form.content.includes(expected)
        : expected.test(this.form.content);

    if (!matches) {
      throw new AssertionError(
        `Expected form content to match ${expected}, but was "${this.form.content}"`,
        expected,
        this.form.content
      );
    }
    return this;
  }

  toHaveButtons(button1: string, button2: string): this {
    if (this.form.button1 !== button1) {
      throw new AssertionError(
        `Expected button1 to be "${button1}", but was "${this.form.button1}"`,
        button1,
        this.form.button1
      );
    }
    if (this.form.button2 !== button2) {
      throw new AssertionError(
        `Expected button2 to be "${button2}", but was "${this.form.button2}"`,
        button2,
        this.form.button2
      );
    }
    return this;
  }

  async clickButton1(): Promise<void> {
    this.player.respondToForm(this.form.id, true);
  }

  async clickButton2(): Promise<void> {
    this.player.respondToForm(this.form.id, false);
  }

  async close(): Promise<void> {
    this.player.closeForm(this.form.id);
  }

  get data(): ModalForm {
    return this.form;
  }
}

export class ActionFormAssertion {
  constructor(
    private form: ActionForm,
    private player: Agent
  ) {}

  toHaveTitle(expected: string | RegExp): this {
    const matches =
      typeof expected === 'string'
        ? this.form.title === expected
        : expected.test(this.form.title);

    if (!matches) {
      throw new AssertionError(
        `Expected form title to match ${expected}, but was "${this.form.title}"`,
        expected,
        this.form.title
      );
    }
    return this;
  }

  toHaveContent(expected: string | RegExp): this {
    const matches =
      typeof expected === 'string'
        ? this.form.content.includes(expected)
        : expected.test(this.form.content);

    if (!matches) {
      throw new AssertionError(
        `Expected form content to match ${expected}, but was "${this.form.content}"`,
        expected,
        this.form.content
      );
    }
    return this;
  }

  toHaveButtonCount(count: number): this {
    if (this.form.buttons.length !== count) {
      throw new AssertionError(
        `Expected ${count} buttons, but found ${this.form.buttons.length}`,
        count,
        this.form.buttons.length
      );
    }
    return this;
  }

  toHaveButton(index: number, text: string | RegExp): this {
    const button = this.form.buttons[index];
    if (!button) {
      throw new AssertionError(
        `Button at index ${index} does not exist`,
        text,
        undefined
      );
    }

    const matches =
      typeof text === 'string' ? button.text === text : text.test(button.text);

    if (!matches) {
      throw new AssertionError(
        `Expected button ${index} to match ${text}, but was "${button.text}"`,
        text,
        button.text
      );
    }
    return this;
  }

  async clickButton(index: number): Promise<void> {
    if (index < 0 || index >= this.form.buttons.length) {
      throw new AssertionError(
        `Button index ${index} out of range (0-${this.form.buttons.length - 1})`,
        `index in range`,
        index
      );
    }
    this.player.respondToForm(this.form.id, index);
  }

  async clickButtonByText(text: string | RegExp): Promise<void> {
    const index = this.form.buttons.findIndex((btn) =>
      typeof text === 'string' ? btn.text === text : text.test(btn.text)
    );

    if (index === -1) {
      throw new AssertionError(
        `No button found matching ${text}`,
        text,
        this.form.buttons.map((b) => b.text)
      );
    }

    this.player.respondToForm(this.form.id, index);
  }

  async close(): Promise<void> {
    this.player.closeForm(this.form.id);
  }

  get data(): ActionForm {
    return this.form;
  }
}

export class CustomFormAssertion {
  constructor(
    private form: CustomForm,
    private player: Agent
  ) {}

  toHaveTitle(expected: string | RegExp): this {
    const matches =
      typeof expected === 'string'
        ? this.form.title === expected
        : expected.test(this.form.title);

    if (!matches) {
      throw new AssertionError(
        `Expected form title to match ${expected}, but was "${this.form.title}"`,
        expected,
        this.form.title
      );
    }
    return this;
  }

  toHaveElementCount(count: number): this {
    if (this.form.content.length !== count) {
      throw new AssertionError(
        `Expected ${count} elements, but found ${this.form.content.length}`,
        count,
        this.form.content.length
      );
    }
    return this;
  }

  toHaveElementAt(
    index: number,
    type: 'label' | 'input' | 'toggle' | 'dropdown' | 'slider' | 'step_slider'
  ): this {
    const element = this.form.content[index];
    if (!element) {
      throw new AssertionError(
        `Element at index ${index} does not exist`,
        type,
        undefined
      );
    }

    if (element.type !== type) {
      throw new AssertionError(
        `Expected element ${index} to be type "${type}", but was "${element.type}"`,
        type,
        element.type
      );
    }
    return this;
  }

  async submit(values: FormResponse): Promise<void> {
    this.player.respondToForm(this.form.id, values);
  }

  async close(): Promise<void> {
    this.player.closeForm(this.form.id);
  }

  get data(): CustomForm {
    return this.form;
  }
}
