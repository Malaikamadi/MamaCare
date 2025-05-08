import React, { useState, useEffect, useRef } from 'react';
import { View, Text, StyleSheet, SafeAreaView, ScrollView, TouchableOpacity, Image, Animated, Dimensions } from 'react-native';
import { Stack, router } from 'expo-router';
import { IconSymbol } from '@/components/ui/IconSymbol';
import { Colors } from '@/constants/Colors';
import { useColorScheme } from '@/hooks/useColorScheme';

// Define interfaces for type safety
interface PregnancyMilestone {
  id: string;
  week: number;
  title: string;
  description: string;
  babySize: string;
  babyWeight: string;
  developmentPoints: string[];
  motherChanges: string[];
  localFoodTip: string;
  warning?: string;
}

interface UpcomingAppointment {
  id: string;
  date: string;
  time: string;
  type: string;
  provider: string;
  location: string;
}

interface HealthTip {
  id: string;
  title: string;
  content: string;
  iconName: string;
}

// Mock pregnancy milestones with Sierra Leone specific information
const PREGNANCY_MILESTONES: Record<number, PregnancyMilestone> = {
  24: {
    id: '24',
    week: 24,
    title: 'Second Trimester',
    description: 'Your baby is growing rapidly and developing facial features.',
    babySize: 'Corn',
    babyWeight: '600g (1.3 lbs)',
    developmentPoints: [
      'Baby can now hear sounds from outside your body',
      'Taste buds are forming',
      'Lung development continues',
      'Skin is still translucent but getting thicker'
    ],
    motherChanges: [
      'You may notice stretch marks appearing',
      'Increased appetite',
      'Back pain may begin',
      'Sleeping may become more difficult'
    ],
    localFoodTip: 'Cassava leaves are rich in vitamins A and C, and folate, which are important for baby\'s development.',
    warning: 'Report any sudden swelling in your hands or face to your CHW immediately.'
  },
  25: {
    id: '25',
    week: 25,
    title: 'Second Trimester',
    description: 'Your baby continues to gain weight and develop more features.',
    babySize: 'Rutabaga',
    babyWeight: '660g (1.5 lbs)',
    developmentPoints: [
      'Brain and lung development continues',
      'Hand grasp reflex developing',
      'Regular movement patterns forming',
      'Hair color and texture developing'
    ],
    motherChanges: [
      'Possible leg cramps',
      'Constipation may be an issue',
      'Snoring may begin or increase',
      'Possible Braxton Hicks contractions'
    ],
    localFoodTip: 'Groundnut soup with fish provides protein and healthy fats for your baby\'s growing brain.',
  },
  26: {
    id: '26',
    week: 26,
    title: 'Second Trimester',
    description: 'Your baby\'s eyes begin to open at this stage.',
    babySize: 'Scallion',
    babyWeight: '760g (1.7 lbs)',
    developmentPoints: [
      'Eyes begin to open',
      'Fingerprints and footprints forming',
      'Immune system developing',
      'Brain wave activity increases'
    ],
    motherChanges: [
      'Swollen ankles more common',
      'Possible headaches',
      'More noticeable fetal movement',
      'Increased vaginal discharge'
    ],
    localFoodTip: 'Okra is a good source of folate, vitamin C, and antioxidants that help with baby\'s development.',
  }
};

// Mock data
const USER_DETAILS = {
  name: 'Aminata',
  pregnancyWeek: 24,
  dueDate: '2025-12-15',
  lastCheckup: '2025-05-01',
  nextCheckup: '2025-05-15',
  healthStatus: 'normal' as 'normal' | 'attention' | 'concern',
  chwName: 'Mary Johnson',
};

const UPCOMING_APPOINTMENTS: UpcomingAppointment[] = [
  {
    id: '1',
    date: '2025-05-15',
    time: '10:00 AM',
    type: 'Prenatal Checkup',
    provider: 'Dr. Sarah Williams',
    location: 'Freetown Community Clinic',
  },
  {
    id: '2',
    date: '2025-05-28',
    time: '2:30 PM',
    type: 'Ultrasound',
    provider: 'Dr. Mohamed Conteh',
    location: 'Sierra Leone Maternal Hospital',
  },
];

// Mock vital monitoring data
const VITAL_SIGNS = {
  heartRate: {
    current: 82,
    min: 75,
    max: 90,
    normal: [70, 95],
    unit: 'bpm',
  },
  bloodPressure: {
    systolic: 118,
    diastolic: 75,
    normal: [{ systolic: 110, diastolic: 70 }, { systolic: 120, diastolic: 80 }],
    unit: 'mmHg',
  },
  temperature: {
    current: 36.7,
    normal: [36.1, 37.2],
    unit: '°C',
  },
  lastUpdated: '10 minutes ago',
};

export default function DashboardScreen(): React.JSX.Element {
  const colorScheme = useColorScheme();
  const [showMilestoneDetails, setShowMilestoneDetails] = useState<boolean>(false);
  const [showMenu, setShowMenu] = useState<boolean>(false);
  const [heartRateData, setHeartRateData] = useState<number[]>([78, 80, 81, 79, 82, 83, 82, 81, 82, 84, 82, 81]);
  const [currentHeartRate, setCurrentHeartRate] = useState<number>(VITAL_SIGNS.heartRate.current);
  const pulseAnim = useRef(new Animated.Value(1)).current;
  
  // Simulate heartbeat pulsing animation
  useEffect(() => {
    const pulseAnimation = Animated.sequence([
      Animated.timing(pulseAnim, {
        toValue: 1.1,
        duration: 300,
        useNativeDriver: true,
      }),
      Animated.timing(pulseAnim, {
        toValue: 1,
        duration: 500,
        useNativeDriver: true,
      }),
    ]);
    
    // Create recurring heartbeat animation
    Animated.loop(
      pulseAnimation,
      { iterations: -1 }
    ).start();
    
    // Simulate changing heart rate data
    const interval = setInterval(() => {
      const newValue = Math.floor(Math.random() * 6) + 78;
      setCurrentHeartRate(newValue);
      setHeartRateData(prevData => {
        return [...prevData.slice(1), newValue];
      });
    }, 2000);
    
    return () => {
      pulseAnim.stopAnimation();
      clearInterval(interval);
    };
  }, []);
  
  // Get current milestone based on pregnancy week
  const getCurrentMilestone = (): PregnancyMilestone => {
    // Get the milestone for current week, or default to week 24 if not found
    return PREGNANCY_MILESTONES[USER_DETAILS.pregnancyWeek] || PREGNANCY_MILESTONES[24];
  };
  
  const currentMilestone = getCurrentMilestone();
  
  // Calculate pregnancy progress
  const calculateProgress = (): number => {
    const totalWeeks = 40;
    return Math.min(100, (USER_DETAILS.pregnancyWeek / totalWeeks) * 100);
  };
  
  // Calculate days until next appointment
  const calculateDaysUntilNextAppointment = (): number => {
    if (!UPCOMING_APPOINTMENTS.length) return 0;
    
    const currentDate = new Date();
    const nextAppointmentDate = new Date(UPCOMING_APPOINTMENTS[0].date);
    const differenceInTime = nextAppointmentDate.getTime() - currentDate.getTime();
    return Math.max(0, Math.ceil(differenceInTime / (1000 * 3600 * 24)));
  };
  
  const daysUntilNextAppointment = calculateDaysUntilNextAppointment();
  
  // Toggle milestone details view
  const toggleMilestoneDetails = (): void => {
    setShowMilestoneDetails(!showMilestoneDetails);
  };
  
  // Toggle hamburger menu
  const toggleMenu = (): void => {
    setShowMenu(!showMenu);
  };
  
  const formatDate = (dateString: string): string => {
    const date = new Date(dateString);
    return date.toLocaleDateString('en-US', { 
      month: 'short', 
      day: 'numeric',
      year: 'numeric'
    });
  };
  
  const renderAppointmentCard = (appointment: UpcomingAppointment): React.JSX.Element => {
    return (
      <TouchableOpacity 
        key={appointment.id}
        style={styles.appointmentCard}
        onPress={() => {
          // Navigate to appointment details in the future
        }}
      >
        <View style={styles.appointmentDateContainer}>
          <View style={styles.appointmentDateCircle}>
            <Text style={styles.appointmentDay}>
              {new Date(appointment.date).getDate()}
            </Text>
            <Text style={styles.appointmentMonth}>
              {new Date(appointment.date).toLocaleDateString('en-US', { month: 'short' })}
            </Text>
          </View>
          <Text style={styles.appointmentTime}>{appointment.time}</Text>
        </View>
        
        <View style={styles.appointmentDetails}>
          <Text style={styles.appointmentType}>{appointment.type}</Text>
          <Text style={styles.appointmentDoctor}>{appointment.provider}</Text>
          <Text style={styles.appointmentLocation}>{appointment.location}</Text>
        </View>
        
        <IconSymbol size={20} name="chevron.right" color="#999" />
      </TouchableOpacity>
    );
  };
  
  // Render heart rate graph points
  const renderHeartRateGraph = (): React.JSX.Element => {
    const width = Dimensions.get('window').width - 120;
    const height = 60;
    const maxValue = 95;
    const minValue = 70;
    
    return (
      <View style={styles.heartRateGraph}>
        {heartRateData.map((value, index) => {
          const x = (width / (heartRateData.length - 1)) * index;
          const normalizedValue = (value - minValue) / (maxValue - minValue);
          const y = height - (normalizedValue * height);
          
          return (
            <View 
              key={`point-${index}`} 
              style={[
                styles.heartRatePoint,
                { 
                  left: x, 
                  top: y,
                  backgroundColor: index === heartRateData.length - 1 ? '#E91E63' : '#fa6b8d'
                }
              ]}
            />
          );
        })}
        
        {/* Connect the dots with lines */}
        {heartRateData.map((value, index) => {
          if (index === 0) return null;
          
          const prevValue = heartRateData[index - 1];
          const currX = (width / (heartRateData.length - 1)) * index;
          const prevX = (width / (heartRateData.length - 1)) * (index - 1);
          const normalizedCurrValue = (value - minValue) / (maxValue - minValue);
          const normalizedPrevValue = (prevValue - minValue) / (maxValue - minValue);
          const currY = height - (normalizedCurrValue * height);
          const prevY = height - (normalizedPrevValue * height);
          
          // Calculate line length and angle
          const length = Math.sqrt(Math.pow(currX - prevX, 2) + Math.pow(currY - prevY, 2));
          const angle = Math.atan2(currY - prevY, currX - prevX) * (180 / Math.PI);
          
          return (
            <View 
              key={`line-${index}`}
              style={[
                styles.heartRateLine,
                {
                  width: length,
                  left: prevX,
                  top: prevY,
                  transform: [{ rotate: `${angle}deg` }],
                  transformOrigin: 'left center',
                }
              ]}
            />
          );
        })}
      </View>
    );
  };
  
  return (
    <SafeAreaView style={styles.container}>
      <Stack.Screen options={{ 
        headerShown: false
      }} />
      
      <View style={styles.topBar}>
        <TouchableOpacity onPress={toggleMenu} style={styles.menuButton}>
          <IconSymbol size={24} name="line.3.horizontal" color="#333" />
        </TouchableOpacity>
        
        <Text style={styles.appTitle}>MamaCare</Text>
        
        <TouchableOpacity 
          style={styles.profileButton}
          onPress={() => {
            router.push('/profile');
          }}
        >
          <Image 
            source={require('@/assets/images/mother.jpeg')} 
            style={styles.profileAvatar} 
            resizeMode="cover"
          />
        </TouchableOpacity>
      </View>
      
      {showMenu && (
        <View style={styles.menuOverlay}>
          <View style={styles.menuContainer}>
            <View style={styles.menuHeader}>
              <Image 
                source={require('@/assets/images/mother.jpeg')} 
                style={styles.menuProfileImage} 
                resizeMode="cover"
              />
              <View style={styles.menuProfileInfo}>
                <Text style={styles.menuProfileName}>{USER_DETAILS.name}</Text>
                <Text style={styles.menuProfileDetails}>{USER_DETAILS.pregnancyWeek} weeks pregnant</Text>
              </View>
            </View>
            
            <View style={styles.menuDivider} />
            
            <TouchableOpacity 
              style={styles.menuItem} 
              onPress={() => {
                setShowMenu(false);
                // Stay on dashboard
              }}
            >
              <IconSymbol size={22} name="house.fill" color="#E91E63" />
              <Text style={[styles.menuItemText, { color: '#E91E63' }]}>Dashboard</Text>
            </TouchableOpacity>
            
            <TouchableOpacity 
              style={styles.menuItem}
              onPress={() => {
                setShowMenu(false);
                router.push('/appointments');
              }}
            >
              <IconSymbol size={22} name="calendar" color="#555" />
              <Text style={styles.menuItemText}>Appointments</Text>
            </TouchableOpacity>
            
            <TouchableOpacity 
              style={styles.menuItem}
              onPress={() => {
                setShowMenu(false);
                router.push('/resources');
              }}
            >
              <IconSymbol size={22} name="book.fill" color="#555" />
              <Text style={styles.menuItemText}>Resources</Text>
            </TouchableOpacity>
            
            <TouchableOpacity 
              style={styles.menuItem}
              onPress={() => {
                setShowMenu(false);
                router.push('/health-records');
              }}
            >
              <IconSymbol size={22} name="doc.text.fill" color="#555" />
              <Text style={styles.menuItemText}>Health Records</Text>
            </TouchableOpacity>
            
            <TouchableOpacity 
              style={styles.menuItem}
              onPress={() => {
                setShowMenu(false);
                router.push('/chat');
              }}
            >
              <IconSymbol size={22} name="message.fill" color="#555" />
              <Text style={styles.menuItemText}>Chat</Text>
            </TouchableOpacity>
            
            <TouchableOpacity 
              style={styles.menuItem}
              onPress={() => {
                setShowMenu(false);
                router.push('/profile');
              }}
            >
              <IconSymbol size={22} name="person.fill" color="#555" />
              <Text style={styles.menuItemText}>Profile</Text>
            </TouchableOpacity>
          </View>
          
          <TouchableOpacity 
            style={styles.menuBackdrop} 
            onPress={() => setShowMenu(false)}
          />
        </View>
      )}
      
      <ScrollView style={styles.scrollView}>
        <View style={styles.monitoringSection}>
          <View style={styles.monitoringHeader}>
            <Text style={styles.monitoringTitle}>Vital Monitoring</Text>
            <Text style={styles.lastUpdated}>Last updated: {VITAL_SIGNS.lastUpdated}</Text>
          </View>
          
          <View style={styles.vitalsRow}>
            <View style={styles.heartRateMonitor}>
              <Animated.View 
                style={[
                  styles.heartIconContainer,
                  { transform: [{ scale: pulseAnim }] }
                ]}
              >
                <IconSymbol size={28} name="heart.fill" color="#E91E63" />
              </Animated.View>
              
              <View style={styles.heartRateContainer}>
                <Text style={styles.heartRateValue}>{currentHeartRate}</Text>
                <Text style={styles.heartRateUnit}>{VITAL_SIGNS.heartRate.unit}</Text>
              </View>
            </View>
            
            <View style={styles.bloodPressureMonitor}>
              <View style={styles.bpIconContainer}>
                <IconSymbol size={24} name="waveform.path.ecg" color="#5E7CE2" />
              </View>
              
              <View style={styles.bpContainer}>
                <Text style={styles.bpValue}>{VITAL_SIGNS.bloodPressure.systolic}/{VITAL_SIGNS.bloodPressure.diastolic}</Text>
                <Text style={styles.bpUnit}>{VITAL_SIGNS.bloodPressure.unit}</Text>
              </View>
            </View>
          </View>
          
          <View style={styles.heartRateGraphContainer}>
            <View style={styles.graphLabels}>
              <Text style={styles.graphLabel}>Heart Rate</Text>
              <View style={styles.graphLegend}>
                <View style={styles.legendItem}>
                  <View style={[styles.legendColor, { backgroundColor: '#fa6b8d' }]} />
                  <Text style={styles.legendText}>Past</Text>
                </View>
                <View style={styles.legendItem}>
                  <View style={[styles.legendColor, { backgroundColor: '#E91E63' }]} />
                  <Text style={styles.legendText}>Current</Text>
                </View>
              </View>
            </View>
            {renderHeartRateGraph()}
          </View>
        </View>
        
        <View style={styles.pregnancyCard}>
          <View style={styles.pregnancyHeader}>
            <Text style={styles.pregnancyTitle}>Your Pregnancy</Text>
            <TouchableOpacity onPress={toggleMilestoneDetails}>
              <Text style={styles.detailsLink}>
                {showMilestoneDetails ? 'Hide Details' : 'Show Details'}
              </Text>
            </TouchableOpacity>
          </View>
          
          <View style={styles.pregnancyInfo}>
            <View style={styles.weekContainer}>
              <Text style={styles.weekNumber}>{USER_DETAILS.pregnancyWeek}</Text>
              <Text style={styles.weekLabel}>Weeks</Text>
            </View>
            
            <View style={styles.progressContainer}>
              <View style={styles.progressDetails}>
                <Text style={styles.progressLabel}>Progress</Text>
                <Text style={styles.dueDate}>Due: {formatDate(USER_DETAILS.dueDate)}</Text>
              </View>
              
              <View style={styles.progressBarContainer}>
                <View 
                  style={[
                    styles.progressBar, 
                    {
                      width: Dimensions.get('window').width * (calculateProgress() / 100)
                    }
                  ]}
                />
              </View>
              
              <View style={styles.babySizeContainer}>
                <Text style={styles.babySizeLabel}>Baby size</Text>
                <Text style={styles.babySizeValue}>{currentMilestone.babySize}</Text>
              </View>
            </View>
          </View>
          
          {showMilestoneDetails && (
            <View style={styles.milestoneDetails}>
              <Text style={styles.milestoneTitle}>{currentMilestone.title}: Week {currentMilestone.week}</Text>
              <Text style={styles.milestoneDescription}>{currentMilestone.description}</Text>
              
              <View style={styles.infoSection}>
                <Text style={styles.infoSectionTitle}>Baby Development</Text>
                {currentMilestone.developmentPoints.map((point, index) => (
                  <View key={`dev-${index}`} style={styles.bulletPoint}>
                    <View style={styles.bullet} />
                    <Text style={styles.bulletText}>{point}</Text>
                  </View>
                ))}
                <Text style={styles.weightText}>Weight: {currentMilestone.babyWeight}</Text>
              </View>
              
              <View style={styles.infoSection}>
                <Text style={styles.infoSectionTitle}>Changes For You</Text>
                {currentMilestone.motherChanges.map((change, index) => (
                  <View key={`change-${index}`} style={styles.bulletPoint}>
                    <View style={styles.bullet} />
                    <Text style={styles.bulletText}>{change}</Text>
                  </View>
                ))}
              </View>
              
              <View style={styles.foodTipContainer}>
                <IconSymbol name="leaf.fill" size={20} color="#33B47B" />
                <Text style={styles.foodTipTitle}>Sierra Leone Food Tip:</Text>
                <Text style={styles.foodTipText}>{currentMilestone.localFoodTip}</Text>
              </View>
              
              {currentMilestone.warning && (
                <View style={styles.warningContainer}>
                  <IconSymbol name="exclamationmark.triangle.fill" size={20} color="#EC5766" />
                  <Text style={styles.warningText}>{currentMilestone.warning}</Text>
                </View>
              )}
            </View>
          )}
        </View>
        
        <View style={styles.nextAppointmentContainer}>
          <Text style={styles.sectionTitle}>Next Appointment</Text>
          
          {UPCOMING_APPOINTMENTS.length > 0 ? (
            <View style={styles.nextAppointmentCard}>
              <View style={styles.nextAppointmentHeader}>
                <View style={styles.nextAppointmentDateContainer}>
                  <IconSymbol size={20} name="calendar" color="#5E7CE2" />
                  <Text style={styles.nextAppointmentDate}>
                    {formatDate(UPCOMING_APPOINTMENTS[0].date)}
                  </Text>
                </View>
                
                <View style={styles.countdownBadge}>
                  <Text style={styles.countdownText}>
                    {daysUntilNextAppointment === 0 ? 'Today' : `In ${daysUntilNextAppointment} days`}
                  </Text>
                </View>
              </View>
              
              <View style={styles.nextAppointmentDetails}>
                <Text style={styles.nextAppointmentType}>
                  {UPCOMING_APPOINTMENTS[0].type}
                </Text>
                <Text style={styles.nextAppointmentTime}>
                  {UPCOMING_APPOINTMENTS[0].time} • {UPCOMING_APPOINTMENTS[0].location}
                </Text>
                <Text style={styles.nextAppointmentDoctor}>
                  {UPCOMING_APPOINTMENTS[0].provider}
                </Text>
              </View>
              
              <View style={styles.appointmentActions}>
                <TouchableOpacity style={styles.rescheduleButton}>
                  <Text style={styles.rescheduleButtonText}>Reschedule</Text>
                </TouchableOpacity>
                
                <TouchableOpacity style={styles.confirmButton}>
                  <Text style={styles.confirmButtonText}>Confirm</Text>
                </TouchableOpacity>
              </View>
            </View>
          ) : (
            <View style={styles.noAppointmentContainer}>
              <IconSymbol size={40} name="calendar.badge.exclamationmark" color="#ccc" />
              <Text style={styles.noAppointmentText}>No upcoming appointments</Text>
              <TouchableOpacity style={styles.scheduleButton}>
                <Text style={styles.scheduleButtonText}>Schedule Now</Text>
              </TouchableOpacity>
            </View>
          )}
        </View>
        
        {UPCOMING_APPOINTMENTS.length > 1 && (
          <View style={styles.upcomingAppointmentsContainer}>
            <View style={styles.sectionHeader}>
              <Text style={styles.sectionTitle}>Upcoming Appointments</Text>
              <TouchableOpacity>
                <Text style={styles.seeAllLink}>See All</Text>
              </TouchableOpacity>
            </View>
            
            {UPCOMING_APPOINTMENTS.slice(1).map(renderAppointmentCard)}
          </View>
        )}
        
        {/* Health tips section removed */}
        
        <View style={styles.chwSection}>
          <Text style={styles.sectionTitle}>Your Community Health Worker</Text>
          <View style={styles.chwCard}>
            <View style={styles.chwInfo}>
              <View style={styles.chwAvatar}>
                <Text style={styles.chwInitials}>
                  {USER_DETAILS.chwName.split(' ').map(n => n[0]).join('')}
                </Text>
              </View>
              <View style={styles.chwDetails}>
                <Text style={styles.chwName}>{USER_DETAILS.chwName}</Text>
                <Text style={styles.chwRole}>Community Health Worker</Text>
              </View>
            </View>
            <TouchableOpacity style={styles.contactChwButton}>
              <IconSymbol size={18} name="message.fill" color="#fff" />
              <Text style={styles.contactChwText}>Contact</Text>
            </TouchableOpacity>
          </View>
        </View>
        
        <View style={styles.sosContainer}>
          <TouchableOpacity style={styles.sosButton}>
            <IconSymbol size={24} name="exclamationmark.triangle.fill" color="#fff" />
            <Text style={styles.sosButtonText}>Emergency SOS</Text>
          </TouchableOpacity>
        </View>
      </ScrollView>
    </SafeAreaView>
  );
}

// Convert numeric percentage to number for width in progress bar
const percentToWidth = (percent: number): number => percent / 100;

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#f8f9fa',
  },
  topBar: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    paddingVertical: 12,
    paddingHorizontal: 16,
    backgroundColor: 'white',
    borderBottomWidth: 1,
    borderBottomColor: '#eee',
    elevation: 2,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.1,
    shadowRadius: 2,
  },
  menuButton: {
    padding: 8,
  },
  appTitle: {
    fontSize: 20,
    fontWeight: 'bold',
    color: '#E91E63',
  },
  profileButton: {
    padding: 8,
  },
  profileAvatar: {
    width: 36,
    height: 36,
    borderRadius: 18,
    backgroundColor: '#E91E63',
  },
  menuOverlay: {
    position: 'absolute',
    top: 0,
    left: 0,
    right: 0,
    bottom: 0,
    zIndex: 100,
    elevation: 5,
  },
  menuBackdrop: {
    position: 'absolute',
    top: 0,
    left: 0,
    right: 0,
    bottom: 0,
    backgroundColor: 'rgba(0,0,0,0.4)',
    zIndex: 1,
  },
  menuContainer: {
    position: 'absolute',
    top: 60, // Adjusted to position below the top bar
    left: 0,
    bottom: 0,
    width: 280,
    backgroundColor: 'white',
    zIndex: 2,
    elevation: 5,
    borderTopRightRadius: 10,
    borderBottomRightRadius: 10,
    shadowColor: '#000',
    shadowOffset: { width: 2, height: 0 },
    shadowOpacity: 0.25,
    shadowRadius: 4,
  },
  menuHeader: {
    padding: 20,
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: '#f8f9fa',
  },
  menuProfileImage: {
    width: 50,
    height: 50,
    borderRadius: 25,
    marginRight: 15,
  },
  menuProfileInfo: {
    flex: 1,
  },
  menuProfileName: {
    fontSize: 18,
    fontWeight: 'bold',
    color: '#333',
  },
  menuProfileDetails: {
    fontSize: 14,
    color: '#666',
    marginTop: 2,
  },
  menuDivider: {
    height: 1,
    backgroundColor: '#eee',
    marginBottom: 10,
  },
  menuItem: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingVertical: 14,
    paddingHorizontal: 24,
  },
  menuItemText: {
    fontSize: 16,
    marginLeft: 16,
    color: '#555',
  },
  monitoringSection: {
    backgroundColor: 'white',
    borderRadius: 12,
    marginBottom: 20,
    padding: 16,
    elevation: 2,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.1,
    shadowRadius: 2,
  },
  monitoringHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 16,
  },
  monitoringTitle: {
    fontSize: 18,
    fontWeight: 'bold',
    color: '#333',
  },
  lastUpdated: {
    fontSize: 12,
    color: '#888',
  },
  vitalsRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    marginBottom: 20,
  },
  heartRateMonitor: {
    flex: 1,
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: '#fff6f7',
    borderRadius: 12,
    padding: 12,
    marginRight: 8,
  },
  heartIconContainer: {
    width: 50,
    height: 50,
    borderRadius: 25,
    backgroundColor: '#ffecef',
    justifyContent: 'center',
    alignItems: 'center',
  },
  heartRateContainer: {
    marginLeft: 12,
  },
  heartRateValue: {
    fontSize: 22,
    fontWeight: 'bold',
    color: '#E91E63',
  },
  heartRateUnit: {
    fontSize: 14,
    color: '#888',
  },
  bloodPressureMonitor: {
    flex: 1,
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: '#f5f7ff',
    borderRadius: 12,
    padding: 12,
    marginLeft: 8,
  },
  bpIconContainer: {
    width: 50,
    height: 50,
    borderRadius: 25,
    backgroundColor: '#e8eeff',
    justifyContent: 'center',
    alignItems: 'center',
  },
  bpContainer: {
    marginLeft: 12,
  },
  bpValue: {
    fontSize: 22,
    fontWeight: 'bold',
    color: '#5E7CE2',
  },
  bpUnit: {
    fontSize: 14,
    color: '#888',
  },
  heartRateGraphContainer: {
    marginTop: 10,
    backgroundColor: '#fff',
    borderRadius: 12,
    padding: 16,
    borderWidth: 1,
    borderColor: '#f0f0f0',
  },
  graphLabels: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 10,
  },
  graphLabel: {
    fontSize: 16,
    fontWeight: '600',
    color: '#333',
  },
  graphLegend: {
    flexDirection: 'row',
  },
  legendItem: {
    flexDirection: 'row',
    alignItems: 'center',
    marginLeft: 12,
  },
  legendColor: {
    width: 8,
    height: 8,
    borderRadius: 4,
    marginRight: 4,
  },
  legendText: {
    fontSize: 12,
    color: '#888',
  },
  heartRateGraph: {
    position: 'relative',
    height: 60,
    marginVertical: 10,
  },
  heartRatePoint: {
    position: 'absolute',
    width: 6,
    height: 6,
    borderRadius: 3,
    backgroundColor: '#E91E63',
    margin: -3,
  },
  heartRateLine: {
    position: 'absolute',
    height: 2,
    backgroundColor: '#fa6b8d',
  },
  milestoneDetails: {
    backgroundColor: '#fff',
    padding: 16,
    marginTop: 12,
    borderTopWidth: 1,
    borderTopColor: '#eee',
    borderRadius: 8,
  },
  milestoneTitle: {
    fontSize: 18,
    fontWeight: 'bold',
    color: '#333',
    marginBottom: 8,
  },
  milestoneDescription: {
    fontSize: 16,
    color: '#555',
    marginBottom: 16,
    lineHeight: 22,
  },
  infoSection: {
    marginBottom: 16,
  },
  infoSectionTitle: {
    fontSize: 16,
    fontWeight: '600',
    color: '#333',
    marginBottom: 8,
  },
  bulletPoint: {
    flexDirection: 'row',
    alignItems: 'flex-start',
    marginBottom: 6,
  },
  bullet: {
    width: 6,
    height: 6,
    borderRadius: 3,
    backgroundColor: '#E91E63',
    marginTop: 7,
    marginRight: 8,
  },
  bulletText: {
    flex: 1,
    fontSize: 14,
    color: '#555',
    lineHeight: 20,
  },
  weightText: {
    fontSize: 14,
    fontWeight: '500',
    color: '#333',
    marginTop: 8,
  },
  foodTipContainer: {
    backgroundColor: '#f1f8f5',
    padding: 12,
    borderRadius: 8,
    marginBottom: 16,
  },
  foodTipTitle: {
    fontSize: 15,
    fontWeight: '600',
    color: '#33B47B',
    marginLeft: 30,
    marginBottom: 4,
    marginTop: -20,
  },
  foodTipText: {
    fontSize: 14,
    color: '#555',
    lineHeight: 20,
  },
  warningContainer: {
    backgroundColor: '#fff6f5',
    padding: 12,
    borderRadius: 8,
    flexDirection: 'row',
    alignItems: 'flex-start',
  },
  warningText: {
    flex: 1,
    fontSize: 14,
    color: '#EC5766',
    marginLeft: 8,
    lineHeight: 20,
  },
  scrollView: {
    flex: 1,
    paddingHorizontal: 16,
  },
  header: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginTop: 10,
    marginBottom: 20,
  },
  greeting: {
    fontSize: 16,
    color: '#777',
  },
  name: {
    fontSize: 24,
    fontWeight: 'bold',
    color: '#333',
  },
  notificationButton: {
    width: 44,
    height: 44,
    borderRadius: 22,
    backgroundColor: '#f0f0f0',
    justifyContent: 'center',
    alignItems: 'center',
  },
  pregnancyCard: {
    backgroundColor: '#fff',
    borderRadius: 12,
    padding: 16,
    marginBottom: 24,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.05,
    shadowRadius: 3,
    elevation: 2,
  },
  pregnancyHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 16,
  },
  pregnancyTitle: {
    fontSize: 18,
    fontWeight: '600',
    color: '#333',
  },
  detailsLink: {
    fontSize: 14,
    color: '#5E7CE2',
  },
  pregnancyInfo: {
    flexDirection: 'row',
    alignItems: 'center',
  },
  weekContainer: {
    alignItems: 'center',
    marginRight: 16,
    backgroundColor: '#E91E63',
    width: 70,
    height: 70,
    borderRadius: 35,
    justifyContent: 'center',
  },
  weekNumber: {
    fontSize: 24,
    fontWeight: 'bold',
    color: '#fff',
  },
  weekLabel: {
    fontSize: 12,
    color: 'rgba(255,255,255,0.8)',
  },
  progressContainer: {
    flex: 1,
  },
  progressDetails: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 8,
  },
  progressLabel: {
    fontSize: 14,
    color: '#777',
  },
  dueDate: {
    fontSize: 14,
    color: '#777',
  },
  progressBarContainer: {
    height: 8,
    backgroundColor: '#f0f0f0',
    borderRadius: 4,
    marginBottom: 8,
  },
  progressBar: {
    height: 8,
    backgroundColor: '#E91E63',
    borderRadius: 4,
  },
  babySizeContainer: {
    flexDirection: 'row',
    alignItems: 'center',
  },
  babySizeLabel: {
    fontSize: 14,
    color: '#777',
    marginRight: 4,
  },
  babySizeValue: {
    fontSize: 14,
    fontWeight: '600',
    color: '#333',
  },
  nextAppointmentContainer: {
    marginBottom: 24,
  },
  sectionTitle: {
    fontSize: 18,
    fontWeight: '600',
    color: '#333',
    marginBottom: 12,
  },
  nextAppointmentCard: {
    backgroundColor: '#fff',
    borderRadius: 12,
    padding: 16,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.05,
    shadowRadius: 3,
    elevation: 2,
  },
  nextAppointmentHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 12,
  },
  nextAppointmentDateContainer: {
    flexDirection: 'row',
    alignItems: 'center',
  },
  nextAppointmentDate: {
    fontSize: 16,
    fontWeight: '600',
    color: '#333',
    marginLeft: 8,
  },
  countdownBadge: {
    backgroundColor: '#E8F0FF',
    paddingHorizontal: 10,
    paddingVertical: 4,
    borderRadius: 12,
  },
  countdownText: {
    fontSize: 12,
    color: '#5E7CE2',
    fontWeight: '500',
  },
  nextAppointmentDetails: {
    marginBottom: 16,
  },
  nextAppointmentType: {
    fontSize: 16,
    fontWeight: '600',
    color: '#333',
    marginBottom: 4,
  },
  nextAppointmentTime: {
    fontSize: 14,
    color: '#555',
    marginBottom: 4,
  },
  nextAppointmentDoctor: {
    fontSize: 14,
    color: '#777',
  },
  appointmentActions: {
    flexDirection: 'row',
    justifyContent: 'flex-end',
  },
  rescheduleButton: {
    paddingHorizontal: 16,
    paddingVertical: 8,
    borderRadius: 4,
    backgroundColor: '#f0f0f0',
    marginRight: 8,
  },
  rescheduleButtonText: {
    fontSize: 14,
    color: '#555',
  },
  confirmButton: {
    paddingHorizontal: 16,
    paddingVertical: 8,
    borderRadius: 4,
    backgroundColor: '#5E7CE2',
  },
  confirmButtonText: {
    fontSize: 14,
    color: '#fff',
  },
  noAppointmentContainer: {
    backgroundColor: '#fff',
    borderRadius: 12,
    padding: 24,
    alignItems: 'center',
    justifyContent: 'center',
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.05,
    shadowRadius: 3,
    elevation: 2,
  },
  noAppointmentText: {
    fontSize: 16,
    color: '#777',
    marginTop: 12,
    marginBottom: 16,
  },
  scheduleButton: {
    paddingHorizontal: 20,
    paddingVertical: 10,
    backgroundColor: '#E91E63',
    borderRadius: 25,
  },
  scheduleButtonText: {
    fontSize: 14,
    color: '#fff',
    fontWeight: '500',
  },
  upcomingAppointmentsContainer: {
    marginBottom: 24,
  },
  sectionHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 12,
  },
  seeAllLink: {
    fontSize: 14,
    color: '#5E7CE2',
  },
  appointmentCard: {
    backgroundColor: '#fff',
    borderRadius: 12,
    padding: 16,
    marginBottom: 12,
    flexDirection: 'row',
    alignItems: 'center',
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.05,
    shadowRadius: 3,
    elevation: 2,
  },
  appointmentDateContainer: {
    alignItems: 'center',
    marginRight: 16,
  },
  appointmentDateCircle: {
    width: 50,
    height: 50,
    borderRadius: 25,
    backgroundColor: '#f0f0f0',
    justifyContent: 'center',
    alignItems: 'center',
    marginBottom: 4,
  },
  appointmentDay: {
    fontSize: 18,
    fontWeight: 'bold',
    color: '#333',
  },
  appointmentMonth: {
    fontSize: 12,
    color: '#777',
  },
  appointmentTime: {
    fontSize: 12,
    color: '#777',
  },
  appointmentDetails: {
    flex: 1,
  },
  appointmentType: {
    fontSize: 16,
    fontWeight: '600',
    color: '#333',
    marginBottom: 4,
  },
  appointmentDoctor: {
    fontSize: 14,
    color: '#555',
    marginBottom: 2,
  },
  appointmentLocation: {
    fontSize: 14,
    color: '#777',
  },
  healthTipsContainer: {
    marginBottom: 24,
  },
  tipCard: {
    backgroundColor: '#fff',
    borderRadius: 12,
    padding: 16,
    marginBottom: 12,
    flexDirection: 'row',
    alignItems: 'center',
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.05,
    shadowRadius: 3,
    elevation: 2,
  },
  tipIconContainer: {
    width: 50,
    height: 50,
    borderRadius: 25,
    backgroundColor: '#5E7CE2',
    justifyContent: 'center',
    alignItems: 'center',
    marginRight: 16,
  },
  tipContent: {
    flex: 1,
  },
  tipTitle: {
    fontSize: 16,
    fontWeight: '600',
    color: '#333',
    marginBottom: 4,
  },
  tipText: {
    fontSize: 14,
    color: '#555',
    lineHeight: 20,
  },
  chwSection: {
    marginBottom: 24,
  },
  chwCard: {
    backgroundColor: '#fff',
    borderRadius: 12,
    padding: 16,
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.05,
    shadowRadius: 3,
    elevation: 2,
  },
  chwInfo: {
    flexDirection: 'row',
    alignItems: 'center',
  },
  chwAvatar: {
    width: 50,
    height: 50,
    borderRadius: 25,
    backgroundColor: '#5E7CE2',
    justifyContent: 'center',
    alignItems: 'center',
    marginRight: 12,
  },
  chwInitials: {
    fontSize: 18,
    fontWeight: 'bold',
    color: '#fff',
  },
  chwDetails: {
    flex: 1,
  },
  chwName: {
    fontSize: 16,
    fontWeight: '600',
    color: '#333',
    marginBottom: 2,
  },
  chwRole: {
    fontSize: 14,
    color: '#777',
  },
  contactChwButton: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: '#5E7CE2',
    paddingHorizontal: 16,
    paddingVertical: 8,
    borderRadius: 20,
  },
  contactChwText: {
    color: '#fff',
    fontSize: 14,
    fontWeight: '500',
    marginLeft: 6,
  },
  sosContainer: {
    marginBottom: 40,
  },
  sosButton: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    backgroundColor: '#EC5766',
    padding: 16,
    borderRadius: 12,
  },
  sosButtonText: {
    color: '#fff',
    fontSize: 16,
    fontWeight: '600',
    marginLeft: 8,
  },
});
