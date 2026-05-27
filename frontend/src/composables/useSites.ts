import { ref } from 'vue'

import {
  createSite,
  deleteCategory as deleteCategoryRequest,
  deleteSite as deleteSiteRequest,
  getCategoryStats,
  getStats,
  listCategories,
  listSites,
  renameCategory as renameCategoryRequest,
  updateSite,
} from '@/lib/api'
import type { CategoryStat, Site, SiteInput, Stats } from '@/types/api'

export function useSites() {
  const sites = ref<Site[]>([])
  const categories = ref<string[]>([])
  const categoryStats = ref<CategoryStat[]>([])
  const stats = ref<Stats>({ siteCount: 0, categoryCount: 0, coverage: '0%' })
  const category = ref('全部')
  const query = ref('')

  async function loadSitesOnly() {
    sites.value = await listSites({ category: category.value, q: query.value })
  }

  async function loadAll() {
    const [nextCategories, nextSites, nextStats] = await Promise.all([
      listCategories(),
      listSites({ category: category.value, q: query.value }),
      getStats(),
    ])
    categories.value = nextCategories
    sites.value = nextSites
    stats.value = nextStats
  }

  async function loadCategoryStats() {
    categoryStats.value = await getCategoryStats()
  }

  async function saveSite(input: SiteInput, id?: string) {
    if (id) {
      await updateSite(id, input)
    } else {
      await createSite(input)
    }
    await loadAll()
  }

  async function removeSite(site: Site) {
    await deleteSiteRequest(site.id)
    await loadAll()
  }

  async function removeCategory(name: string) {
    await deleteCategoryRequest(name)
    if (category.value === name) category.value = '全部'
    await loadAll()
    await loadCategoryStats()
  }

  async function renameCategory(name: string, nextName: string) {
    const result = await renameCategoryRequest(name, nextName)
    if (category.value === name) category.value = result.name
    await loadAll()
    await loadCategoryStats()
  }

  return {
    sites,
    categories,
    categoryStats,
    stats,
    category,
    query,
    loadAll,
    loadSitesOnly,
    loadCategoryStats,
    saveSite,
    removeSite,
    removeCategory,
    renameCategory,
  }
}
