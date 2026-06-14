import { z } from "zod";

const strongPasswordRegex =
  /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[!"#$%&'()*+,\-./:;<=>?@[\\\]^_{|}~]).{8,72}$/;

export const passwordStrengthHint =
  "Use 8 a 72 caracteres com letra maiuscula, letra minuscula, numero e caractere especial.";

export const passwordStrengthMessage =
  "A senha deve ter 8 a 72 caracteres, incluindo letra maiuscula, letra minuscula, numero e caractere especial.";

export function createStrongPasswordSchema(fieldLabel: string) {
  return z
    .string()
    .min(8, `${fieldLabel} deve ter pelo menos 8 caracteres`)
    .max(72, `${fieldLabel} deve ter no maximo 72 caracteres`)
    .regex(strongPasswordRegex, passwordStrengthMessage);
}
