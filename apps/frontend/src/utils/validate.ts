import { z } from 'zod'

export const CreateWalletSchema = z
  .object({
    name: z.string().min(2, 'Name must be at least 2 characters').max(50, 'Name must be less than 50 characters').trim(),
    description: z.string().max(200, 'Description must be less than 200 characters').optional().default(''),
    initial_balance: z.number().min(0, 'Initial balance cannot be negative'),
    minimum_limit: z.number().min(0, 'Minimum limit cannot be negative')
  })
  .refine((data) => data.minimum_limit <= data.initial_balance, {
    message: 'Minimum limit cannot exceed initial balance',
    path: ['minimum_limit']
  })

export const UpdateWalletSchema = z.object({
  name: z.string().min(2, 'Name must be at least 2 characters').max(50, 'Name must be less than 50 characters').trim(),
  description: z.string().max(200, 'Description must be less than 200 characters').optional().default(''),
  minimum_limit: z.number().min(0, 'Minimum limit cannot be negative')
})

export const CreateTransactionSchema = z.object({
  wallet_id: z.string().optional().default(''),
  type: z.enum(['income', 'expense', 'allocation']),
  amount: z.number().positive('Amount must be greater than 0'),
  category: z.string().max(50).optional().default(''),
  description: z.string().max(500, 'Description cannot exceed 500 characters').optional().default('')
})

export const UpdateTransactionSchema = z.object({
  wallet_id: z.string().optional().default(''),
  type: z.enum(['income', 'expense', 'allocation']),
  amount: z.number().positive('Amount must be greater than 0'),
  description: z.string().max(500, 'Description cannot exceed 500 characters').optional().default('')
})

export const CreateProposalSchema = z.object({
  wallet_id: z.string().min(1, 'Please select a target wallet'),
  title: z.string().min(3, 'Title must be at least 3 characters').max(100, 'Title cannot exceed 100 characters').trim(),
  amount: z.number().positive('Amount must be greater than 0'),
  description: z.string().min(3, 'Description must be at least 3 characters').max(1000, 'Description cannot exceed 1000 characters').trim(),
  request_type: z.enum(['add_transaction', 'edit_transaction', 'delete_transaction']).default('add_transaction'),
  target_transaction_id: z.string().optional(),
  payload: z.any().optional()
})

export const AllocateFundsSchema = z.object({
  wallet_id: z.string().min(1, 'Please select a target wallet'),
  amount: z.number().positive('Allocation amount must be greater than 0'),
  description: z.string().max(200).optional().default('')
})

export const UpdateFamilyNameSchema = z.object({
  name: z.string().min(2, 'Family name must be at least 2 characters').max(100, 'Family name cannot exceed 100 characters').trim()
})

export const FamilySettingsSchema = z.object({
  monthly_income: z.number().min(0, 'Monthly income cannot be negative')
})

export function validateForm<T>(
  schema: z.ZodSchema<T>,
  data: unknown
): { success: true; data: T } | { success: false; errors: Record<string, string> } {
  const result = schema.safeParse(data)
  if (result.success) {
    return { success: true, data: result.data }
  }

  const errors: Record<string, string> = {}
  const issues = (result.error as any).issues || (result.error as any).errors || []
  issues.forEach((err: any) => {
    const path = err.path ? err.path.join('.') : 'root'
    if (!errors[path]) {
      errors[path] = err.message || 'Invalid value'
    }
  })
  return { success: false, errors }
}
