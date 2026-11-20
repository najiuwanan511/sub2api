/**
 * Redeem code API endpoints
 * Handles redeem code redemption for users
 */

import { apiClient } from './client'
import type { PaginatedResponse, RedeemCodeRequest } from '@/types'

export interface WalletResetCredit {
  redeem_code_id: number
  group_id: number | null
  group_name: string
  amount_usd: number
  remaining_usd: number
  expires_at: string
  status: 'active' | 'spent' | 'expired'
}

export interface WalletResetPreview {
  group_name: string
  percent: number
  eligible_usd: number
  credit_usd: number
  validity_days: number
}

export interface RedeemHistoryItem {
  id: number
  code: string
  type: string
  value: number
  status: string
  used_at: string
  created_at: string
  // Notes from admin for admin_balance/admin_concurrency types
  notes?: string
  // Subscription-specific fields
  group_id?: number
  validity_days?: number
  group?: {
    id: number
    name: string
  }
}

/**
 * Redeem a code
 * @param code - Redeem code string
 * @returns Redemption result with updated balance or concurrency
 */
export async function redeem(code: string): Promise<{
  message: string
  type: string
  value: number
  new_balance?: number
  new_concurrency?: number
  reset_credit?: WalletResetCredit
}> {
  const payload: RedeemCodeRequest = { code }

  const { data } = await apiClient.post<{
    message: string
    type: string
    value: number
    new_balance?: number
    new_concurrency?: number
    reset_credit?: WalletResetCredit
  }>('/redeem', payload)

  return data
}

/**
 * Get user's redemption history
 * @returns The requested page of redeemed codes and the total count
 */
export async function getHistory(page = 1, pageSize = 20): Promise<PaginatedResponse<RedeemHistoryItem>> {
  const { data } = await apiClient.get<PaginatedResponse<RedeemHistoryItem>>('/redeem/history', {
    params: { page, page_size: pageSize }
  })
  return data
}

async function previewResetCard(code: string): Promise<WalletResetPreview> {
  const { data } = await apiClient.post<WalletResetPreview>('/redeem/reset-card/preview', { code })
  return data
}

async function applyResetCard(code: string): Promise<WalletResetCredit> {
  const { data } = await apiClient.post<WalletResetCredit>('/redeem/reset-card/apply', { code })
  return data
}

interface WalletResetCreditSummary {
  active_usd: number
  credits: WalletResetCredit[]
}

async function getResetCredits(): Promise<WalletResetCreditSummary> {
  const { data } = await apiClient.get<WalletResetCreditSummary>('/redeem/reset-card/credits')
  return data
}

export const redeemAPI = {
  previewResetCard,
  applyResetCard,
  getResetCredits,
  redeem,
  getHistory
}

export default redeemAPI
