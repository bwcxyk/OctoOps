export function validatePasswordComplexity(password: string) {
  if ([...password].length < 8) {
    return { result: false, message: '密码长度至少为8位' };
  }

  let hasLower = false;
  let hasUpper = false;
  let hasDigit = false;
  let hasSpecial = false;

  for (const char of password) {
    if (/[a-z]/.test(char)) {
      hasLower = true;
    } else if (/[A-Z]/.test(char)) {
      hasUpper = true;
    } else if (/\d/.test(char)) {
      hasDigit = true;
    } else {
      hasSpecial = true;
    }
  }

  const categories = [hasLower, hasUpper, hasDigit, hasSpecial].filter(Boolean).length;
  if (categories < 3) {
    return { result: false, message: '密码需包含数字、大写字母、小写字母、特殊字符中的至少三种' };
  }

  return true;
}

