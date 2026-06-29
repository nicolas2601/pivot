import { apiFetch } from './client';
import {
  SummaryReportSchema,
  ByCategoryReportSchema,
  ByAccountReportSchema,
  MonthlyTrendReportSchema,
  CashflowReportSchema,
  ReportFiltersSchema,
  type SummaryReport,
  type ByCategoryReport,
  type ByAccountReport,
  type MonthlyTrendReport,
  type CashflowReport,
  type ReportFilters
} from '$lib/schemas/report';

export type {
  SummaryReport,
  ByCategoryReport,
  ByAccountReport,
  MonthlyTrendReport,
  CashflowReport,
  ReportFilters
} from '$lib/schemas/report';

function buildQuery(filters: ReportFilters): string {
  const params = new URLSearchParams();
  for (const [key, value] of Object.entries(filters)) {
    if (value !== undefined && value !== null && value !== '') {
      params.append(key, String(value));
    }
  }
  const qs = params.toString();
  return qs ? `?${qs}` : '';
}

export async function getSummary(filters: ReportFilters = {}): Promise<SummaryReport> {
  const validated = ReportFiltersSchema.parse(filters);
  const res = await apiFetch<unknown>(`/reports/summary${buildQuery(validated)}`);
  return SummaryReportSchema.parse(res);
}

export async function getByCategory(filters: ReportFilters = {}): Promise<ByCategoryReport> {
  const validated = ReportFiltersSchema.parse(filters);
  const res = await apiFetch<unknown>(`/reports/by-category${buildQuery(validated)}`);
  return ByCategoryReportSchema.parse(res);
}

export async function getByAccount(filters: ReportFilters = {}): Promise<ByAccountReport> {
  const validated = ReportFiltersSchema.parse(filters);
  const res = await apiFetch<unknown>(`/reports/by-account${buildQuery(validated)}`);
  return ByAccountReportSchema.parse(res);
}

export async function getMonthlyTrend(filters: ReportFilters = {}): Promise<MonthlyTrendReport> {
  const validated = ReportFiltersSchema.parse(filters);
  const res = await apiFetch<unknown>(`/reports/monthly-trend${buildQuery(validated)}`);
  return MonthlyTrendReportSchema.parse(res);
}

export async function getCashflow(filters: ReportFilters = {}): Promise<CashflowReport> {
  const validated = ReportFiltersSchema.parse(filters);
  const res = await apiFetch<unknown>(`/reports/cashflow${buildQuery(validated)}`);
  return CashflowReportSchema.parse(res);
}