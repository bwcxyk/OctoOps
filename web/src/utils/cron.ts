import { CronExpressionParser } from 'cron-parser';

export function isValidCronExpression(expr: string) {
  const value = (expr || '').trim();
  if (!value) return false;
  try {
    CronExpressionParser.parse(value);
    return true;
  } catch {
    return false;
  }
}
