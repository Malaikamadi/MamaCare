import React from 'react';
import { View, Text, StyleSheet, SafeAreaView, ScrollView, TouchableOpacity, Image } from 'react-native';
import { Stack } from 'expo-router';
import { IconSymbol } from '@/components/ui/IconSymbol';
import { Colors } from '@/constants/Colors';
import { useColorScheme } from '@/hooks/useColorScheme';

// Define interfaces for type safety
interface ResourceCategory {
  id: string;
  title: string;
  description: string;
  icon: string;
}

interface ResourceArticle {
  id: string;
  title: string;
  summary: string;
  categoryId: string;
  imageUrl: string;
  readTimeMinutes: number;
}

// Mock data for resource categories
const RESOURCE_CATEGORIES: ResourceCategory[] = [
  {
    id: 'pregnancy',
    title: 'Pregnancy Basics',
    description: 'Essential information for expectant mothers',
    icon: 'figure.pregnant',
  },
  {
    id: 'nutrition',
    title: 'Nutrition & Diet',
    description: 'Healthy eating for you and your baby',
    icon: 'mug.fill',
  },
  {
    id: 'health',
    title: 'Health & Wellness',
    description: 'Staying healthy during pregnancy',
    icon: 'heart.fill',
  },
  {
    id: 'preparations',
    title: 'Birth Preparations',
    description: 'Getting ready for delivery day',
    icon: 'bed.double.fill',
  },
  {
    id: 'baby',
    title: 'Baby Care',
    description: 'Caring for your newborn',
    icon: 'figure.and.child.holdinghands',
  },
];

// Mock data for featured articles
const FEATURED_ARTICLES: ResourceArticle[] = [
  {
    id: '1',
    title: 'Understanding Your Pregnancy Journey in Sierra Leone',
    summary: 'A comprehensive guide to pregnancy care specific to Sierra Leone healthcare system.',
    categoryId: 'pregnancy',
    imageUrl: 'https://via.placeholder.com/300x200',
    readTimeMinutes: 8,
  },
  {
    id: '2',
    title: 'Local Foods for a Healthy Pregnancy',
    summary: 'Nutritional guide featuring foods commonly available in Sierra Leone markets.',
    categoryId: 'nutrition',
    imageUrl: 'https://via.placeholder.com/300x200',
    readTimeMinutes: 5,
  },
  {
    id: '3',
    title: 'Signs of Labor: When to Get Help',
    summary: 'Learn the important signs that indicate labor has begun and when to seek medical attention.',
    categoryId: 'health',
    imageUrl: 'https://via.placeholder.com/300x200',
    readTimeMinutes: 7,
  },
];

// Mock data for recommended articles
const RECOMMENDED_ARTICLES: ResourceArticle[] = [
  {
    id: '4',
    title: 'Preparing Your Home for a Newborn',
    summary: 'Simple steps to make your home safe and comfortable for your new baby.',
    categoryId: 'preparations',
    imageUrl: 'https://via.placeholder.com/300x200',
    readTimeMinutes: 6,
  },
  {
    id: '5',
    title: 'Common Pregnancy Discomforts and Solutions',
    summary: 'Practical advice for managing common pregnancy discomforts naturally.',
    categoryId: 'health',
    imageUrl: 'https://via.placeholder.com/300x200',
    readTimeMinutes: 10,
  },
  {
    id: '6',
    title: 'Breastfeeding Basics for New Mothers',
    summary: 'A guide to successful breastfeeding for first-time mothers.',
    categoryId: 'baby',
    imageUrl: 'https://via.placeholder.com/300x200',
    readTimeMinutes: 8,
  },
];

export default function ResourcesScreen(): React.JSX.Element {
  const colorScheme = useColorScheme();
  const [searchQuery, setSearchQuery] = React.useState<string>('');

  const renderCategoryCard = (category: ResourceCategory): React.JSX.Element => {
    return (
      <TouchableOpacity
        key={category.id}
        style={styles.categoryCard}
        onPress={() => {
          // Navigate to category details in the future
        }}
      >
        <View style={styles.categoryIcon}>
          <IconSymbol size={24} name={category.icon} color="#fff" />
        </View>
        <View style={styles.categoryContent}>
          <Text style={styles.categoryTitle}>{category.title}</Text>
          <Text style={styles.categoryDescription}>{category.description}</Text>
        </View>
        <IconSymbol size={20} name="chevron.right" color="#999" />
      </TouchableOpacity>
    );
  };

  const renderArticleCard = (article: ResourceArticle): React.JSX.Element => {
    return (
      <TouchableOpacity
        key={article.id}
        style={styles.articleCard}
        onPress={() => {
          // Navigate to article details in the future
        }}
      >
        <Image 
          source={{ uri: article.imageUrl }}
          style={styles.articleImage}
        />
        <View style={styles.articleContent}>
          <Text style={styles.articleTitle}>{article.title}</Text>
          <Text style={styles.articleSummary} numberOfLines={2}>{article.summary}</Text>
          <View style={styles.articleMeta}>
            <IconSymbol size={14} name="clock.fill" color="#999" />
            <Text style={styles.readTime}>{article.readTimeMinutes} min read</Text>
          </View>
        </View>
      </TouchableOpacity>
    );
  };

  return (
    <SafeAreaView style={styles.container}>
      <Stack.Screen options={{ title: 'Resources' }} />
      
      <View style={styles.header}>
        <Text style={styles.headerTitle}>Resources</Text>
        <TouchableOpacity style={styles.searchButton}>
          <IconSymbol size={22} name="magnifyingglass" color="#555" />
        </TouchableOpacity>
      </View>
      
      <ScrollView style={styles.scrollView}>
        <View style={styles.section}>
          <Text style={styles.sectionTitle}>Categories</Text>
          <View style={styles.categoriesContainer}>
            {RESOURCE_CATEGORIES.map(renderCategoryCard)}
          </View>
        </View>
        
        <View style={styles.section}>
          <Text style={styles.sectionTitle}>Featured Articles</Text>
          <ScrollView 
            horizontal
            showsHorizontalScrollIndicator={false}
            style={styles.horizontalScrollView}
            contentContainerStyle={styles.horizontalScrollContainer}
          >
            {FEATURED_ARTICLES.map(renderArticleCard)}
          </ScrollView>
        </View>
        
        <View style={styles.section}>
          <Text style={styles.sectionTitle}>Recommended For You</Text>
          <View style={styles.articlesContainer}>
            {RECOMMENDED_ARTICLES.map(renderArticleCard)}
          </View>
        </View>
        
        <View style={styles.videoSection}>
          <View style={styles.videoSectionHeader}>
            <Text style={styles.sectionTitle}>Video Resources</Text>
            <TouchableOpacity>
              <Text style={styles.seeAllLink}>See All</Text>
            </TouchableOpacity>
          </View>
          <TouchableOpacity style={styles.featuredVideo}>
            <View style={styles.videoThumbnail}>
              <IconSymbol size={40} name="play.circle.fill" color="#fff" />
            </View>
            <View style={styles.videoContent}>
              <Text style={styles.videoTitle}>Preparing for a Safe Delivery</Text>
              <Text style={styles.videoDuration}>12:45</Text>
            </View>
          </TouchableOpacity>
        </View>
      </ScrollView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#f8f9fa',
  },
  header: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingHorizontal: 16,
    paddingVertical: 16,
  },
  headerTitle: {
    fontSize: 28,
    fontWeight: 'bold',
    color: '#333',
  },
  searchButton: {
    backgroundColor: '#f0f0f0',
    width: 44,
    height: 44,
    borderRadius: 22,
    justifyContent: 'center',
    alignItems: 'center',
  },
  scrollView: {
    flex: 1,
  },
  section: {
    marginBottom: 24,
  },
  sectionTitle: {
    fontSize: 18,
    fontWeight: '600',
    color: '#333',
    marginHorizontal: 16,
    marginBottom: 12,
  },
  categoriesContainer: {
    paddingHorizontal: 16,
  },
  categoryCard: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: '#fff',
    borderRadius: 12,
    padding: 16,
    marginBottom: 10,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.05,
    shadowRadius: 3,
    elevation: 2,
  },
  categoryIcon: {
    width: 44,
    height: 44,
    borderRadius: 22,
    backgroundColor: '#E91E63',
    justifyContent: 'center',
    alignItems: 'center',
    marginRight: 16,
  },
  categoryContent: {
    flex: 1,
  },
  categoryTitle: {
    fontSize: 16,
    fontWeight: '600',
    color: '#333',
    marginBottom: 4,
  },
  categoryDescription: {
    fontSize: 13,
    color: '#777',
  },
  horizontalScrollView: {
    paddingVertical: 8,
  },
  horizontalScrollContainer: {
    paddingHorizontal: 16,
  },
  articleCard: {
    backgroundColor: '#fff',
    borderRadius: 12,
    overflow: 'hidden',
    width: 280,
    marginRight: 12,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.1,
    shadowRadius: 3,
    elevation: 2,
  },
  articleImage: {
    width: '100%',
    height: 150,
    resizeMode: 'cover',
  },
  articleContent: {
    padding: 14,
  },
  articleTitle: {
    fontSize: 16,
    fontWeight: '600',
    color: '#333',
    marginBottom: 6,
  },
  articleSummary: {
    fontSize: 13,
    color: '#666',
    marginBottom: 8,
    lineHeight: 18,
  },
  articleMeta: {
    flexDirection: 'row',
    alignItems: 'center',
  },
  readTime: {
    fontSize: 12,
    color: '#999',
    marginLeft: 4,
  },
  articlesContainer: {
    paddingHorizontal: 16,
  },
  videoSection: {
    marginBottom: 32,
  },
  videoSectionHeader: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    marginHorizontal: 16,
    marginBottom: 12,
  },
  seeAllLink: {
    fontSize: 14,
    color: '#E91E63',
  },
  featuredVideo: {
    marginHorizontal: 16,
    backgroundColor: '#000',
    borderRadius: 12,
    overflow: 'hidden',
  },
  videoThumbnail: {
    height: 180,
    backgroundColor: '#333',
    justifyContent: 'center',
    alignItems: 'center',
  },
  videoContent: {
    padding: 16,
    backgroundColor: '#222',
  },
  videoTitle: {
    fontSize: 16,
    fontWeight: '600',
    color: '#fff',
    marginBottom: 6,
  },
  videoDuration: {
    fontSize: 13,
    color: '#aaa',
  },
});
