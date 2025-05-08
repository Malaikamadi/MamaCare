import React, { useState, useRef, useEffect } from 'react';
import { 
  View, 
  Text, 
  StyleSheet, 
  SafeAreaView, 
  TextInput, 
  TouchableOpacity, 
  FlatList, 
  KeyboardAvoidingView, 
  Platform,
  Image,
  Switch
} from 'react-native';
import { Stack } from 'expo-router';
import { IconSymbol } from '@/components/ui/IconSymbol';
import { Colors } from '@/constants/Colors';
import { useColorScheme } from '@/hooks/useColorScheme';

// Define types for our messages and participants
interface Message {
  id: string;
  text: string;
  sender: 'user' | 'bot' | 'nurse';
  timestamp: Date;
}

interface ChatParticipant {
  id: string;
  name: string;
  avatar: string;
  type: 'ai' | 'nurse';
  isOnline?: boolean;
}

// Mock nurse data
const nurses: ChatParticipant[] = [
  {
    id: '1',
    name: 'Nurse Maria',
    avatar: 'nurse',  // We'll use the local nurse.jpg image
    type: 'nurse',
    isOnline: true
  },
  {
    id: '2',
    name: 'Nurse James',
    avatar: 'nurse',  // We'll use the same nurse image for now
    type: 'nurse',
    isOnline: false
  }
];

// AI bot participant
const aiBot: ChatParticipant = {
  id: 'ai-1',
  name: 'MamaCare Assistant',
  avatar: 'ai-bot',  // We'll use a locally imported image in the render function
  type: 'ai',
  isOnline: true
};

// Mock user data - in a real app, this would come from user authentication
const user = {
  id: 'user-1',
  name: 'Fatima',
  avatar: 'mother',  // We'll use the mother.jpeg from assets
};

// Sample initial messages for the AI bot
const initialBotMessages: Message[] = [
  {
    id: 'bot-1',
    text: `Hello ${user.name}! I'm your MamaCare AI assistant. How are you feeling today?`,
    sender: 'bot',
    timestamp: new Date(Date.now() - 24 * 60 * 60 * 1000) // 1 day ago
  }
];

// Sample initial messages for the nurse
const initialNurseMessages: Message[] = [
  {
    id: 'nurse-1',
    text: `Hi ${user.name}, this is Nurse Maria. I'm here to answer any questions about your pregnancy. How can I help you today?`,
    sender: 'nurse',
    timestamp: new Date(Date.now() - 2 * 24 * 60 * 60 * 1000) // 2 days ago
  }
];

// Bot responses based on user input
const botResponses: Record<string, string[]> = {
  greeting: [
    `Hi ${user.name}! How are you feeling today?`, 
    `Hello ${user.name}! Is there anything I can help you with?`,
    `Good day ${user.name}! How's your pregnancy journey going?`
  ],
  feeling: [
    "I'm sorry to hear that you're not feeling well. Can you tell me more about your symptoms?",
    "It's good to hear you're feeling well! Remember to stay hydrated and take your prenatal vitamins.",
    "Pregnancy can be challenging sometimes. Have you discussed these feelings with your healthcare provider?"
  ],
  question: [
    "That's a great question! Based on your stage of pregnancy, it's normal to experience these changes.",
    "I'd recommend discussing this with your healthcare provider at your next appointment.",
    "Many mothers have similar questions. The important thing is to listen to your body and consult with professionals."
  ],
  default: [
    "I'm here to support you throughout your pregnancy journey.",
    "Is there anything specific about your pregnancy that you'd like to discuss?",
    "Remember, every pregnancy is unique. Don't hesitate to reach out if you have any concerns."
  ]
};

export default function ChatScreen(): React.JSX.Element {
  const colorScheme = useColorScheme();
  const [messages, setMessages] = useState<Message[]>([]);
  const [inputText, setInputText] = useState<string>('');
  const [isChatWithNurse, setIsChatWithNurse] = useState<boolean>(false);
  const [selectedNurse, setSelectedNurse] = useState<ChatParticipant>(nurses[0]);
  const flatListRef = useRef<FlatList>(null);

  // Initialize messages based on selected chat type
  useEffect(() => {
    if (isChatWithNurse) {
      setMessages(initialNurseMessages);
    } else {
      setMessages(initialBotMessages);
    }
  }, [isChatWithNurse]);

  // Auto scroll to bottom when new messages are added
  useEffect(() => {
    if (flatListRef.current && messages.length > 0) {
      flatListRef.current.scrollToEnd({ animated: true });
    }
  }, [messages]);

  // Helper function to categorize user message
  const categorizeMessage = (text: string): string => {
    const lowerText = text.toLowerCase();
    if (lowerText.includes('hi') || lowerText.includes('hello') || lowerText.includes('hey')) {
      return 'greeting';
    } else if (lowerText.includes('feel') || lowerText.includes('feeling') || lowerText.includes('sick') || lowerText.includes('pain')) {
      return 'feeling';
    } else if (lowerText.includes('?') || lowerText.includes('what') || lowerText.includes('how') || lowerText.includes('why')) {
      return 'question';
    }
    return 'default';
  };

  // Generate a random response based on message category
  const getRandomResponse = (category: string): string => {
    const responses = botResponses[category] || botResponses.default;
    return responses[Math.floor(Math.random() * responses.length)];
  };

  // Send message
  const handleSend = (): void => {
    if (inputText.trim() === '') return;

    // Add user message
    const newUserMessage: Message = {
      id: `user-${Date.now()}`,
      text: inputText,
      sender: 'user',
      timestamp: new Date()
    };
    
    const updatedMessages = [...messages, newUserMessage];
    setMessages(updatedMessages);
    setInputText('');

    // Simulate response after a brief delay
    setTimeout(() => {
      const category = categorizeMessage(inputText);
      const responseText = getRandomResponse(category);
      
      const responseMessage: Message = {
        id: `${isChatWithNurse ? 'nurse' : 'bot'}-${Date.now()}`,
        text: responseText,
        sender: isChatWithNurse ? 'nurse' : 'bot',
        timestamp: new Date()
      };
      
      setMessages(prevMessages => [...prevMessages, responseMessage]);
    }, 1000);
  };

  // Toggle between AI chatbot and nurse
  const toggleChatType = (): void => {
    setIsChatWithNurse(!isChatWithNurse);
  };

  // Format timestamp
  const formatTime = (date: Date): string => {
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  };

  // Render a chat message
  const renderMessage = ({ item }: { item: Message }): React.JSX.Element => {
    const isUserMessage = item.sender === 'user';
    
    // Helper function to get the correct avatar image
    const getAvatarSource = (avatarType: string) => {
      switch(avatarType) {
        case 'mother':
          return require('@/assets/images/mother.jpeg');
        case 'nurse':
          return require('@/assets/images/nurse.jpeg');
        case 'ai-bot':
          return { uri: 'https://ui-avatars.com/api/?name=MC&background=E91E63&color=fff' };
        default:
          return { uri: 'https://ui-avatars.com/api/?name=MC&background=E91E63&color=fff' };
      }
    };
    
    return (
      <View style={[
        styles.messageContainer,
        isUserMessage ? styles.userMessageContainer : styles.botMessageContainer
      ]}>
        {!isUserMessage && (
          <Image 
            source={getAvatarSource(isChatWithNurse ? selectedNurse.avatar : aiBot.avatar)} 
            style={styles.avatar} 
          />
        )}
        <View style={[
          styles.messageBubble,
          isUserMessage ? styles.userMessageBubble : styles.botMessageBubble
        ]}>
          <Text style={[styles.messageText, isUserMessage && styles.userMessageText]}>{item.text}</Text>
          <Text style={styles.timestamp}>{formatTime(item.timestamp)}</Text>
        </View>
        {isUserMessage && (
          <Image 
            source={getAvatarSource(user.avatar)} 
            style={styles.avatar} 
          />
        )}
      </View>
    );
  };

  return (
    <SafeAreaView style={styles.container}>
      <Stack.Screen options={{ headerShown: false }} />
      
      <View style={styles.header}>
        <View style={styles.headerContent}>
          <View style={styles.profileContainer}>
            <Image 
              source={isChatWithNurse 
                ? require('@/assets/images/nurse.jpeg')
                : { uri: 'https://ui-avatars.com/api/?name=MC&background=E91E63&color=fff' }
              } 
              style={styles.headerAvatar} 
            />
            <View>
              <Text style={styles.headerTitle}>
                {isChatWithNurse ? selectedNurse.name : aiBot.name}
              </Text>
              <Text style={styles.headerSubtitle}>
                {isChatWithNurse ? 'Healthcare Professional' : 'AI Assistant'}
              </Text>
            </View>
          </View>
          <View style={styles.toggleContainer}>
            <Text style={styles.toggleLabel}>AI</Text>
            <Switch
              value={isChatWithNurse}
              onValueChange={toggleChatType}
              trackColor={{ false: '#E91E63', true: '#4CAF50' }}
              thumbColor={isChatWithNurse ? '#fff' : '#fff'}
            />
            <Text style={styles.toggleLabel}>Nurse</Text>
          </View>
        </View>
      </View>
      
      <FlatList
        ref={flatListRef}
        data={messages}
        renderItem={renderMessage}
        keyExtractor={item => item.id}
        style={styles.chatList}
        contentContainerStyle={styles.chatContent}
      />
      
      <KeyboardAvoidingView
        behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
        keyboardVerticalOffset={100}
        style={styles.inputContainer}
      >
        <TextInput
          style={styles.input}
          value={inputText}
          onChangeText={setInputText}
          placeholder="Type a message..."
          placeholderTextColor="#888"
          multiline
        />
        <TouchableOpacity 
          style={styles.sendButton} 
          onPress={handleSend}
          disabled={inputText.trim() === ''}
        >
          <IconSymbol size={24} name="arrow.up.circle.fill" color="#fff" />
        </TouchableOpacity>
      </KeyboardAvoidingView>
      
      {isChatWithNurse && (
        <View style={styles.disclaimer}>
          <Text style={styles.disclaimerText}>
            Nurse response times may vary. For emergencies, please call your healthcare provider directly.
          </Text>
        </View>
      )}
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#f8f9fa',
  },
  header: {
    backgroundColor: '#E91E63',
    paddingTop: 50,
    paddingBottom: 15,
    paddingHorizontal: 16,
    borderBottomLeftRadius: 20,
    borderBottomRightRadius: 20,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.2,
    shadowRadius: 4,
    elevation: 5,
  },
  headerContent: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  profileContainer: {
    flexDirection: 'row',
    alignItems: 'center',
  },
  headerAvatar: {
    width: 44,
    height: 44,
    borderRadius: 22,
    marginRight: 12,
    backgroundColor: '#f1f1f1',
    borderWidth: 2,
    borderColor: '#fff',
  },
  headerTitle: {
    fontSize: 18,
    fontWeight: 'bold',
    color: '#fff',
  },
  headerSubtitle: {
    fontSize: 12,
    color: 'rgba(255, 255, 255, 0.8)',
  },
  toggleContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: 'rgba(255, 255, 255, 0.2)',
    borderRadius: 20,
    padding: 5,
    marginTop: 5,
  },
  toggleLabel: {
    color: '#fff',
    marginHorizontal: 5,
    fontSize: 12,
  },
  chatList: {
    flex: 1,
    padding: 16,
  },
  chatContent: {
    paddingBottom: 16,
  },
  messageContainer: {
    flexDirection: 'row',
    marginBottom: 16,
    alignItems: 'flex-end',
  },
  userMessageContainer: {
    justifyContent: 'flex-end',
  },
  botMessageContainer: {
    justifyContent: 'flex-start',
  },
  avatar: {
    width: 36,
    height: 36,
    borderRadius: 18,
    marginHorizontal: 8,
    backgroundColor: '#f1f1f1',
  },
  messageBubble: {
    maxWidth: '70%',
    borderRadius: 20,
    padding: 12,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.1,
    shadowRadius: 2,
    elevation: 2,
  },
  userMessageBubble: {
    backgroundColor: '#E91E63',
    borderBottomRightRadius: 0,
  },
  botMessageBubble: {
    backgroundColor: '#fff',
    borderBottomLeftRadius: 0,
  },
  messageText: {
    fontSize: 16,
    color: '#333',
  },
  userMessageText: {
    color: '#fff',
  },
  timestamp: {
    fontSize: 10,
    color: '#888',
    alignSelf: 'flex-end',
    marginTop: 4,
  },
  inputContainer: {
    flexDirection: 'row',
    padding: 12,
    backgroundColor: '#fff',
    borderTopWidth: 1,
    borderTopColor: '#eee',
    alignItems: 'center',
  },
  input: {
    flex: 1,
    backgroundColor: '#f0f0f0',
    borderRadius: 20,
    paddingHorizontal: 16,
    paddingVertical: 10,
    maxHeight: 120,
    fontSize: 16,
  },
  sendButton: {
    marginLeft: 12,
    backgroundColor: '#E91E63',
    width: 40,
    height: 40,
    borderRadius: 20,
    justifyContent: 'center',
    alignItems: 'center',
  },
  disclaimer: {
    padding: 8,
    backgroundColor: 'rgba(255, 193, 7, 0.2)',
    borderRadius: 8,
    margin: 8,
  },
  disclaimerText: {
    fontSize: 12,
    color: '#856404',
    textAlign: 'center',
  }
});
