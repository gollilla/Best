export interface BaseForm {
  id: number;
  title: string;
}

export interface ModalForm extends BaseForm {
  type: 'modal';
  content: string;
  button1: string;
  button2: string;
}

export interface ActionFormButton {
  text: string;
  image?: {
    type: 'path' | 'url';
    data: string;
  };
}

export interface ActionForm extends BaseForm {
  type: 'action';
  content: string;
  buttons: ActionFormButton[];
}

export type CustomFormElement =
  | CustomFormLabel
  | CustomFormInput
  | CustomFormToggle
  | CustomFormDropdown
  | CustomFormSlider
  | CustomFormStepSlider;

export interface CustomFormLabel {
  type: 'label';
  text: string;
}

export interface CustomFormInput {
  type: 'input';
  text: string;
  placeholder?: string;
  default?: string;
}

export interface CustomFormToggle {
  type: 'toggle';
  text: string;
  default?: boolean;
}

export interface CustomFormDropdown {
  type: 'dropdown';
  text: string;
  options: string[];
  default?: number;
}

export interface CustomFormSlider {
  type: 'slider';
  text: string;
  min: number;
  max: number;
  step?: number;
  default?: number;
}

export interface CustomFormStepSlider {
  type: 'step_slider';
  text: string;
  steps: string[];
  default?: number;
}

export interface CustomForm extends BaseForm {
  type: 'form';
  content: CustomFormElement[];
}

export type Form = ModalForm | ActionForm | CustomForm;

export type FormResponse = boolean | number | null | (string | number | boolean)[];
